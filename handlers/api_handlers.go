package handlers

import (
	"fmt"
	goccy "github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/models"
	"github.com/sharify-labs/spine/services"
	"github.com/sharify-labs/spine/validators"
	"net/http"
	"strings"
)

func ZephyrProxy(c echo.Context) error {
	user := c.Get("user").(models.AuthorizedUser)
	return clients.HTTP.ForwardToZephyr(c, user.ZephyrJWT)
}

func ResetToken(c echo.Context) error {
	var token *services.ZephyrToken
	var err error

	user := c.Get("user").(models.AuthorizedUser)
	if token, err = services.NewZephyrToken(user.ID); err != nil {
		// TODO: These are repeated in ProvideConfig() handler. Prob should make 1 func.
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

// ListAvailableDomains returns a JSON array of all available root domain names.
func ListAvailableDomains(c echo.Context) error {
	if domains, err := clients.HTTP.GetOrFetchAvailableDomains(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	} else {
		return c.JSON(http.StatusOK, domains)
	}
}

// ListHosts returns a JSON array of all hosts registered by a given user.
func ListHosts(c echo.Context) error {
	user := c.Get("user").(models.AuthorizedUser)
	hostnames, err := database.GetAllHostnames(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, hostnames)
}

// CreateHost creates new hosts for a user.
// Root domain must be registered first. This can be checked with ListDomainsHandler.
func CreateHost(c echo.Context) error {
	sub := validators.SanitizeSubdomain(c.FormValue("subDomain"))
	root := c.FormValue("rootDomain") // TODO: Make sure rootDomain came from the list of available domains and is not being injected some other way.
	user := c.Get("user").(models.AuthorizedUser)

	host := services.NewHostFromParts(sub, root, user.ID)
	if host == nil {
		// Hostname does not meet format requirements. Should prob be validated on frontend too.
		return echo.NewHTTPError(http.StatusBadRequest, "Hostname format must be sub.root.tld")
	}

	// Check if root domain is in list of available domains
	availableDomains, err := clients.HTTP.GetOrFetchAvailableDomains()
	if err != nil {
		clients.Sentry.CaptureErr(c, fmt.Errorf("failed to fetch available domains: %v", err))
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if _, ok := availableDomains[host.Root]; !ok {
		clients.Sentry.CaptureErr(c, fmt.Errorf("user (%s) tried create host (%s) but root missing from available domains", user.ID, host.Full))
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// Publish host (add to Cloudflare & Database)
	err = host.Register()
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
	user := c.Get("user").(models.AuthorizedUser)

	if host := services.NewHostFromFull(hostname, user.ID); host != nil {
		if err := host.Delete(); err != nil {
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

	user := c.Get("user").(models.AuthorizedUser)
	token, err := services.NewZephyrToken(user.ID)
	if err != nil {
		clients.Sentry.CaptureErr(c, fmt.Errorf("failed to generate zephyr token: %v", err))
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	cfg.Headers.UploadToken = token.Value
	// TODO: Prompt users when generating config if they want to be prompted for custom paths or upload lifetimes
	cfg.Arguments.Secret = "{prompt:Enter custom secret or press OK to skip|}"
	cfg.Arguments.Duration = "{prompt:Enter number of hours until upload expires or skip for permanent|}"

	hostnames, err := database.GetAllHostnames(user.ID)
	if err != nil {
		clients.Sentry.CaptureErr(c, err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	switch len(hostnames) {
	case 0:
		cfg.Arguments.Host = config.HostDefault
	case 1:
		cfg.Arguments.Host = hostnames[0]
	default:
		// TODO: Make it optional for users to select "randomize" from the menu when generating config
		// 		 In those cases, replace 'select' with 'random'
		cfg.Arguments.Host = fmt.Sprintf("{select:%s}", strings.Join(hostnames, "|"))
	}

	userConfigContent, err := goccy.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+user.ID+".sxcu")

	return c.Blob(http.StatusOK, echo.MIMEOctetStream, userConfigContent)
}
