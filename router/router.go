package router

import (
	"github.com/labstack/echo/v4"
	"github.com/posty/spine/handlers"
	mw "github.com/posty/spine/middleware"
	"net/http"
)

const (
	RouteHome  string = "/"
	RouteLogin string = "/login"

	RouteDiscordAuth         string = "/auth/discord"
	RouteDiscordAuthCallback string = "/auth/discord/callback"

	RouteDashboard   string = "/dashboard"
	RouteListDomains string = "/api/list-domains"

	RouteResetKey   string = "/api/reset-key/:user_id"
	RouteListHosts  string = "/api/list-hosts/:user_id"
	RouteCreateHost string = "/api/create-host/:user_id/:hostname"
	RouteDeleteHost string = "/api/delete-host/:user_id/:hostname"
)

func Setup(e *echo.Echo) {
	e.GET(RouteHome, func(c echo.Context) error {
		return c.Redirect(http.StatusFound, RouteLogin)
	})
	e.GET(RouteLogin, handlers.LoginHandler)
	e.GET(RouteDiscordAuth, handlers.DiscordAuthHandler)
	e.GET(RouteDiscordAuthCallback, handlers.DiscordAuthCallbackHandler)

	// Protected routes
	e.GET(RouteDashboard, handlers.DashboardHandler, mw.IsAuthenticated)
	e.GET(RouteListDomains, handlers.ListDomainsHandler, mw.IsAuthenticated)
	e.GET(RouteResetKey, handlers.ResetKeyHandler, mw.IsAuthenticated) // This could be POST in the future. GET allows for easy testing.
	e.GET(RouteListHosts, handlers.ListHostsHandler, mw.IsAuthenticated)
	e.GET(RouteCreateHost, handlers.CreateHostHandler, mw.IsAuthenticated) // This could be POST in the future. GET allows for easy testing.
	e.GET(RouteDeleteHost, handlers.DeleteHostHandler, mw.IsAuthenticated)
}
