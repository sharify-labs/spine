package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/models"
	"net/http"
)

func Root(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/dashboard")
}

func Login(c echo.Context) error {
	return c.HTML(http.StatusOK, `<a href="/auth/discord">Login with Discord</a>`)
}

type DashboardData struct {
	Username string
	UserID   string
	Domains  []string
	Hosts    []HostData
}
type HostData struct {
	Name string
}

func DisplayDashboard(c echo.Context) error {
	domains, err := clients.Cloudflare.AvailableDomains()
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	user := c.Get("user").(models.AuthorizedUser)
	hostnames, err := database.GetAllHostnames(user.ID)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	var hosts []HostData
	for _, h := range hostnames {
		hosts = append(hosts, HostData{Name: h})
	}

	return c.Render(
		http.StatusOK, "dashboard.html",
		DashboardData{
			Username: user.Discord.Username,
			UserID:   user.ID,
			Domains:  domains,
			Hosts:    hosts,
		},
	)
}
