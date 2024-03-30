package router

import (
	"github.com/labstack/echo/v4"
	h "github.com/sharify-labs/spine/handlers"
	mw "github.com/sharify-labs/spine/middleware"
)

// Setup initializes all routes:
//
// Root:
// GET     /       -> handlers.Root
// GET     /login  -> handlers.Login
//
// Auth:
// GET     /auth/discord           -> handlers.DiscordAuth
// GET     /auth/discord/callback  -> handlers.DiscordAuthCallback
//
// Protected:
// GET     /dashboard       	-> handlers.DisplayDashboard
// GET     /api/reset-token 	-> handlers.ResetToken
// GET     /api/gallery     	-> handlers.DisplayGallery
// GET	   /api/config/:type    -> handlers.ProvideConfig  // :type must be files/pastes/redirects
//
// GET	   /api/domains      -> handlers.ListAvailableDomains
//
// GET     /api/hosts        -> handlers.ListHosts
// POST    /api/hosts        -> handlers.CreateHost
// DELETE  /api/hosts/:name  -> handlers.DeleteHost
func Setup(e *echo.Echo) {
	// Root routes
	e.GET("", h.Root)
	e.GET("/login", h.Login)

	// Auth routes
	auth := e.Group("/auth")
	{
		discord := auth.Group("/discord")
		{
			discord.GET("", h.DiscordAuth)
			discord.GET("/callback", h.DiscordAuthCallback)
		}
	}

	// Protected routes
	e.GET("/dashboard", h.DisplayDashboard, mw.RequireSession)
	api := e.Group("/api", mw.RequireSession)
	{
		api.GET("/reset-token", h.ResetToken)
		//api.GET("/gallery", h.DisplayGallery)
		api.GET("/config/:type", h.ProvideConfig)

		domains := api.Group("/domains")
		{
			domains.GET("", h.ListAvailableDomains)
		}

		hosts := api.Group("/hosts")
		{
			hosts.GET("", h.ListHosts)
			hosts.POST("", h.CreateHost)
			hosts.DELETE("/:name", h.DeleteHost)
		}
	}
}
