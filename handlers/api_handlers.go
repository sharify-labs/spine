package handlers

import (
	"bytes"
	"fmt"
	goccy "github.com/goccy/go-json"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/models"
	"github.com/sharify-labs/spine/services"
	"github.com/sharify-labs/spine/utils"
	"io"
	"net/http"
	"strings"
)

func getUserFromSession(c echo.Context) models.AuthorizedUser {
	sess, _ := session.Get("session", c)
	return sess.Values["auth_user"].(models.AuthorizedUser)
}

func ResetToken(c echo.Context) error {
	var token *services.ZephyrToken
	var err error

	user := getUserFromSession(c)
	if token, err = services.NewZephyrToken(user.ID); err != nil {
		// TODO: These are repeated in ProvideConfig() handler. Prob should make 1 func.
		c.Logger().Error(err)
		clients.Sentry.CaptureErr(c, fmt.Errorf("failed to generate zephyr token: %v", err))
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.HTML(
		http.StatusOK,
		`<pre style="margin: 0; font-size: 16px; background-color: #131516; color: #ccc; border: 1px solid #ccc; padding: 0;">
		<code id="token">`+token.Value+`</code></pre>
		<button onclick="copyContent('token')">Copy Token</button>`,
	)
}

func DisplayGallery(c echo.Context) error {
	user := getUserFromSession(c)
	uploads, err := database.GetUserUploads(user.ID)
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
	// TODO: dont read API Key from env every time
	url := "https://ejl.me/api/uploads/" + user.ID + "?key=" + config.Str("CANVAS_API_KEY")
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
	user := getUserFromSession(c)
	hostnames, err := database.GetAllHostnames(user.ID)
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
	user := getUserFromSession(c)

	host := services.NewHostDTO(hostname, user.ID)
	if host == nil {
		// Hostname does not meet format requirements. Should prob be validated on frontend too.
		return echo.NewHTTPError(http.StatusBadRequest, "Hostname format must be sub.root.tld")
	}

	// Publish host (add to Cloudflare & Database)
	err := host.Register()
	if err != nil {
		clients.Sentry.CaptureErr(c, fmt.Errorf("failed to register host(%s, %s, %s): %v", user.ID, host.Sub, host.Root, err))
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true, // TODO: Replace this
	})
}

func DeleteHost(c echo.Context) error {
	hostname := c.Param("name")
	user := getUserFromSession(c)

	if host := services.NewHostDTO(hostname, user.ID); host != nil {
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
// Note: It also regenerates their upload token.
func ProvideConfig(c echo.Context) error {
	cfg := models.NewShareXConfig(strings.ToLower(c.Param("type")))
	if cfg == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	user := getUserFromSession(c)
	token, err := services.NewZephyrToken(user.ID)
	if err != nil {
		c.Logger().Error(err)
		clients.Sentry.CaptureErr(c, fmt.Errorf("failed to generate zephyr token: %v", err))
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	cfg.Headers.UploadToken = token.Value
	// TODO: Prompt users when generating config if they want to be prompted for custom paths or upload lifetimes
	cfg.Parameters.CustomPath = "{prompt:Enter custom path or press OK to skip|}"
	cfg.Parameters.MaxHours = "{prompt:Enter number of hours until upload expires or press OK for permanent|}"

	hostnames, err := database.GetAllHostnames(user.ID)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	switch len(hostnames) {
	case 0:
		cfg.Parameters.Host = config.HostDefault
	case 1:
		cfg.Parameters.Host = hostnames[0]
	default:
		// TODO: Make it optional for users to select "randomize" from the menu when generating config
		// 		 In those cases, replace 'select' with 'random'
		cfg.Parameters.Host = fmt.Sprintf("{select:%s}", strings.Join(hostnames, "|"))
	}

	userConfigContent, err := goccy.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+user.ID+".sxcu")

	return c.Blob(http.StatusOK, echo.MIMEOctetStream, userConfigContent)
}
