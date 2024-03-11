package handlers

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/posty/spine/config"
	"github.com/posty/spine/database"
	"github.com/posty/spine/dto"
	"github.com/posty/spine/security"
	"net/http"
	"strings"
)

func DashboardHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	username := sess.Values["discord_username"]
	if username == nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	return c.HTML(http.StatusOK, `<div>Welcome to the dashboard, `+username.(string)+`!</div><form action="/generate-key" method="POST"><button type="submit">Generate Key</button></form>`)
}

// ResetKeyHandler refreshes an existing user's upload-key.
func ResetKeyHandler(c echo.Context) error {
	userID := strings.TrimSpace(c.Param("user_id"))
	if userID == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	key := security.NewZephyrKey()
	if key == nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	if err := database.UpdateUserKey(userID, key.Hash, key.Salt); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, echo.Map{
		config.UserHeader: userID,
		config.KeyHeader:  key.Key,
	})
}

// ListDomainsHandler returns a JSON array of all available root domain names.
func ListDomainsHandler(c echo.Context) error {
	domains, err := database.GetDomainsAvailable()
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	var names []string
	for _, domain := range domains {
		names = append(names, domain.Name)
	}
	return c.JSON(http.StatusOK, names)
}

// ListHostsHandler returns a JSON array of all hosts registered by a given user.
func ListHostsHandler(c echo.Context) error {
	userID := strings.TrimSpace(c.Param("user_id"))
	if userID == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	hosts, err := database.GetAllHosts(userID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	var names []string
	for _, host := range hosts {
		names = append(names, host.Full)
	}
	return c.JSON(http.StatusOK, names)
}

// CreateHostHandler creates new hosts for a user.
// Root domain of hostname must be registered first. This can be checked with ListDomainsHandler.
func CreateHostHandler(c echo.Context) error {
	hostname := c.Param("hostname")
	host := dto.NewHost(hostname)
	if host == nil {
		// Hostname does not meet format requirements. Should prob be validated on frontend too.
		return c.NoContent(http.StatusBadRequest)
	}

	userID := strings.TrimSpace(c.Param("user_id"))
	err := database.InsertHost(userID, host.Sub, host.Root)
	if err != nil {
		// Unable to create host in DB. Maybe root domain doesn't exist, or user doesn't exist, or something else.
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, echo.Map{
		config.UserHeader: userID,
		config.HostHeader: host.Full,
	})
}

// DeleteHostHandler deletes a registered hostname.
func DeleteHostHandler(c echo.Context) error {
	hostname := strings.TrimSpace(c.Param("hostname"))
	if hostname == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	userID := strings.TrimSpace(c.Param("user_id"))
	err := database.DeleteHost(userID, hostname)
	if err != nil {
		// Unable to create host in DB. Maybe root domain doesn't exist, or user doesn't exist, or something else.
		return c.NoContent(http.StatusBadRequest)
	}
	return c.NoContent(http.StatusOK)
}
