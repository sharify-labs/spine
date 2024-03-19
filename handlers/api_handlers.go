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
	"github.com/sharify-labs/spine/security"
	"github.com/sharify-labs/spine/services"
	"github.com/sharify-labs/spine/utils"
	"io"
	"net/http"
)

func getUserID(c echo.Context) string {
	sess, _ := session.Get("session", c)
	return sess.Values["user_id"].(string)
}

// ResetToken refreshes an existing user's upload-key.
func ResetToken(c echo.Context) error {
	userID := getUserID(c)
	token := security.NewZephyrToken()
	if token == nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	if err := database.UpdateUserToken(userID, token.Hash, token.Salt); err != nil {
		return c.NoContent(http.StatusInternalServerError)
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
	if userID == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	uploads, err := database.GetUserUploads(userID)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	var storeKeys []string
	for _, u := range uploads {
		storeKeys = append(storeKeys, u.StorageKey)
	}

	rqBody, err := goccy.Marshal(storeKeys)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	url := "https://ejl.me/api/uploads/" + userID + "?key=" + config.GetStr("CANVAS_API_KEY")
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(rqBody))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	var base64Images []string
	if err = goccy.Unmarshal(respBody, &base64Images); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, base64Images)
}

// ListAvailableDomains returns a JSON array of all available root domain names.
// Domains are fetched from Cloudflare on each request.
// TODO: Cache this in Redis (12-24 hours sounds reasonable)
func ListAvailableDomains(c echo.Context) error {
	if domains, err := clients.Cloudflare.AvailableDomains(); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	} else {
		return c.JSON(http.StatusOK, domains)
	}
}

// ListHosts returns a JSON array of all hosts registered by a given user.
func ListHosts(c echo.Context) error {
	userID := getUserID(c)
	hostnames, err := database.GetAllHostnames(userID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, hostnames)
}

// CreateHost creates new hosts for a user.
// Root domain must be registered first. This can be checked with ListDomainsHandler.
func CreateHost(c echo.Context) error {
	hostname := utils.CompileHostname(c.FormValue("subDomain"), c.FormValue("rootDomain"))
	userID := getUserID(c)

	host := services.NewHostDTO(hostname, userID)
	if host == nil {
		// Hostname does not meet format requirements. Should prob be validated on frontend too.
		return c.NoContent(http.StatusBadRequest)
	}

	// Publish host (add to Cloudflare & Database)
	err := host.Register()
	if err != nil {
		clients.Sentry.CaptureErr(c, fmt.Errorf("failed to register host(%s, %s, %s): %v", userID, host.Sub, host.Root, err))
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, echo.Map{
		config.UserHeader: userID,
		config.HostHeader: host.Full,
	})
}

// DeleteHost deletes a registered hostname.
func DeleteHost(c echo.Context) error {
	hostname := c.Param("name")
	userID := getUserID(c)

	if host := services.NewHostDTO(hostname, userID); host != nil {
		if err := host.Delete(); err != nil {
			c.Logger().Error(err)
			clients.Sentry.CaptureErr(c, fmt.Errorf("error deleting host: %v", err))
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	return c.NoContent(http.StatusBadRequest)
}
