package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/posty/spine/handlers"
)

const (
	RouteHome  string = "/"
	RouteLogin string = "/login"

	RouteDiscordAuth         string = "/auth/discord"
	RouteDiscordAuthCallback string = "/auth/discord/callback"

	RouteListDomains string = "/api/list-domains"

	RouteResetKey   string = "/api/reset-key/:user_id"
	RouteListHosts  string = "/api/list-hosts/:user_id"
	RouteCreateHost string = "/api/create-host/:user_id/:hostname"
	RouteDeleteHost string = "/api/delete-host/:user_id/:hostname"
)

func Setup(a *fiber.App) {
	a.Get(RouteHome, func(c *fiber.Ctx) error {
		return c.Redirect(RouteLogin, fiber.StatusFound)
	})
	a.Get(RouteLogin, handlers.LoginHandler)
	a.Get(RouteDiscordAuth, handlers.DiscordAuthHandler)
	a.Get(RouteDiscordAuthCallback, handlers.DiscordAuthCallbackHandler)
	a.Get(RouteListDomains, handlers.ListDomainsHandler)
	a.Get(RouteResetKey, handlers.ResetKeyHandler) // This could be POST in the future. GET allows for easy testing.
	a.Get(RouteListHosts, handlers.ListHostsHandler)
	a.Get(RouteCreateHost, handlers.CreateHostHandler) // This could be POST in the future. GET allows for easy testing.
	a.Get(RouteDeleteHost, handlers.DeleteHostHandler)
	a.Use(func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotFound) })
}
