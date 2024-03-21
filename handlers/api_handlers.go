package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	goccy "github.com/goccy/go-json"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/models"
	"github.com/sharify-labs/spine/security"
	"github.com/sharify-labs/spine/services"
	"github.com/sharify-labs/spine/utils"
	"io"
	"net/http"
)

func getUserID(c echo.Context) string {
	sess, _ := session.Get("session", c)
	// TODO: In hindsight, this doesn't seem safe
	return sess.Values["user_id"].(string)
}

func ResetToken(c echo.Context) error {
	var token *security.ZephyrToken
	var err error

	userID := getUserID(c)

	if token, err = security.NewZephyrToken(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if err = token.AssignToUser(userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.HTML(
		http.StatusOK,
		`<pre style="margin: 0; font-size: 16px; background-color: #131516; color: #ccc; border: 1px solid #ccc; padding: 0;">
		<code id="token">`+token.Value+`</code></pre>
		<button onclick="copyContent('token')">Copy Token</button>`,
	)
}

func DisplayGallery(c echo.Context) error {
	userID := getUserID(c)
	uploads, err := database.GetUserUploads(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	var storeKeys []string
	for _, u := range uploads {
		storeKeys = append(storeKeys, u.StorageKey)
	}

	rqBody, err := goccy.Marshal(storeKeys)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	url := "https://ejl.me/api/uploads/" + userID + "?key=" + config.GetStr("CANVAS_API_KEY")
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(rqBody))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	var base64Images []string
	if err = goccy.Unmarshal(respBody, &base64Images); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, base64Images)
}

// ListAvailableDomains returns a JSON array of all available root domain names.
// Domains are fetched from Cloudflare on each request.
// TODO: Cache this in Redis (12-24 hours sounds reasonable)
func ListAvailableDomains(c echo.Context) error {
	if domains, err := clients.Cloudflare.AvailableDomains(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	} else {
		return c.JSON(http.StatusOK, domains)
	}
}

// ListHosts returns a JSON array of all hosts registered by a given user.
func ListHosts(c echo.Context) error {
	userID := getUserID(c)
	hostnames, err := database.GetAllHostnames(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, hostnames)
}

// CreateHost creates new hosts for a user.
// Root domain must be registered first. This can be checked with ListDomainsHandler.
func CreateHost(c echo.Context) error {
	sub := utils.SanitizeSubdomain(c.FormValue("subDomain"))
	hostname := utils.CompileHostname(sub, c.FormValue("rootDomain")) // TODO: Make sure rootDomain came from the list of available domains and is not being injected some other way.
	userID := getUserID(c)

	host := services.NewHostDTO(hostname, userID)
	if host == nil {
		// Hostname does not meet format requirements. Should prob be validated on frontend too.
		return echo.NewHTTPError(http.StatusBadRequest, "Hostname format must be sub.root.tld")
	}

	// Publish host (add to Cloudflare & Database)
	err := host.Register()
	if err != nil {
		clients.Sentry.CaptureErr(c, fmt.Errorf("failed to register host(%s, %s, %s): %v", userID, host.Sub, host.Root, err))
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		config.UserHeader: userID,
		config.HostHeader: host.Full,
	})
}

func DeleteHost(c echo.Context) error {
	hostname := c.Param("name")
	userID := getUserID(c)

	if host := services.NewHostDTO(hostname, userID); host != nil {
		if err := host.Delete(); err != nil {
			c.Logger().Error(err)
			clients.Sentry.CaptureErr(c, fmt.Errorf("error deleting host: %v", err))
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		return c.NoContent(http.StatusOK)
	}

	return echo.NewHTTPError(http.StatusInternalServerError, "Hostname format must be sub.root.tld")
}

// ProvideConfig returns a ShareX config file for the user.
func ProvideConfig(c echo.Context) error {
	userId := getUserID(c)

	userConfig := &models.ShareXConfig{
		Version:         "4.3.0",
		Name:            "Sharify (Image)",
		DestinationType: "ImageUploader, TextUploader, FileUploader",
		RequestMethod:   "POST",
		RequestURL:      "https://xericl.dev/api/v1/share",
		Parameters: models.Parameters{
			Host:       "CHANGE ME",
			CustomPath: "CHANGE ME",
			MaxHours:   0,
		},
		Headers: models.Headers{
			UploadUser:  userId,
			UploadToken: "CHANGE ME",
		},
		Body:         "MultipartFormData",
		FileFormName: "file",
	}
	if err = token.AssignToUser(userId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	uConfig.Headers.UploadToken = token.Value

	userConfigContent, err := json.Marshal(userConfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+userId+".sxcu")
	c.Response().Header().Set(echo.HeaderContentType, "application/octet-stream")

	return c.Blob(http.StatusOK, "application/octet-stream", userConfigContent)
}
