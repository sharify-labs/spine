package handlers

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/posty/spine/clients"
	"github.com/posty/spine/database"
	"net/http"
)

func Root(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/login")
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
	sess, _ := session.Get("session", c)

	username, ok := sess.Values["discord_username"].(string)
	if !ok {
		return c.Redirect(http.StatusFound, "/login")
	}

	domains, err := clients.Cloudflare.AvailableDomains()
	if err != nil {
		c.Logger().Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	userID := getUserID(c)
	hostnames, err := database.GetAllHostnames(userID)
	if err != nil {
		c.Logger().Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var hosts []HostData
	for _, h := range hostnames {
		hosts = append(hosts, HostData{Name: h})
	}

	return c.Render(
		http.StatusOK, "dashboard.html",
		DashboardData{
			Username: username,
			UserID:   userID,
			Domains:  domains,
			Hosts:    hosts,
		},
	)
}
