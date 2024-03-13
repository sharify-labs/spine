package handlers

import (
	"bytes"
	"github.com/goccy/go-json"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/posty/spine/config"
	"github.com/posty/spine/database"
	"github.com/posty/spine/dto"
	"github.com/posty/spine/security"
	"github.com/posty/spine/services"
	"io"
	"net/http"
)

func getUserID(c echo.Context) string {
	sess, _ := session.Get("session", c)
	return sess.Values["user_id"].(string)
}

// ResetTokenHandler refreshes an existing user's upload-key.
func ResetTokenHandler(c echo.Context) error {
	userID := getUserID(c)
	token := security.NewZephyrToken()
	if token == nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	if err := database.UpdateUserToken(userID, token.Hash, token.Salt); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, echo.Map{
		config.UserHeader:  userID,
		config.TokenHeader: token.Value,
	})
}

func GalleryHandler(c echo.Context) error {
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

	rqBody, err := json.Marshal(storeKeys)
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
	if err = json.Unmarshal(respBody, &base64Images); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, base64Images)
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
	userID := getUserID(c)
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

func CreateDomainHandler(c echo.Context) error {
	domain := c.Param("name")
	err := database.InsertDomain(domain)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// CreateHostHandler creates new hosts for a user.
// Root domain must be registered first. This can be checked with ListDomainsHandler.
func CreateHostHandler(c echo.Context) error {
	hostname := c.Param("name")
	host := dto.NewHost(hostname)
	if host == nil {
		// Hostname does not meet format requirements. Should prob be validated on frontend too.
		return c.NoContent(http.StatusBadRequest)
	}

	userID := getUserID(c)
	// Add to Cloudflare
	err := services.CreateCNAME(userID, host.Sub, host.Root)
	if err != nil {
		c.Logger().Errorf("failed to create CNAME(%s, %s, %s): %v", userID, host.Sub, host.Root, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	// Add to Database
	err = database.InsertHost(userID, host.Sub, host.Root)
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
	hostname := c.Param("name")
	userID := getUserID(c)
	err := database.DeleteHost(userID, hostname)
	if err != nil {
		// Unable to create host in DB. Maybe root domain doesn't exist, or user doesn't exist, or something else.
		return c.NoContent(http.StatusBadRequest)
	}
	return c.NoContent(http.StatusOK)
}
