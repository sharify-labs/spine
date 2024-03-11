package router

import (
	"github.com/labstack/echo/v4"
	"github.com/posty/spine/handlers"
	mw "github.com/posty/spine/middleware"
)

// Setup initializes all routes:
//
// Root:
// GET     / -> handlers.RootHandler
// GET     /login -> handlers.LoginHandler
//
// Auth:
// GET     /auth/discord -> handlers.DiscordAuthHandler
// GET     /auth/discord/callback -> handlers.DiscordAuthCallbackHandler
//
// Protected:
// GET     /dashboard -> handlers.DashboardHandler
// GET     /api/reset-token -> handlers.ResetTokenHandler
// GET     /api/gallery -> handlers.GalleryHandler
//
// GET     /api/domains -> handlers.ListDomainsHandler
// POST    /api/domains/:name -> handlers.CreateDomainHandler
//
// GET     /api/hosts -> handlers.ListHostsHandler
// POST    /api/hosts/:name -> handlers.CreateHostHandler
// DELETE  /api/hosts/:name -> handlers.DeleteHostHandler
func Setup(e *echo.Echo) {
	// Root routes
	e.GET("/", handlers.RootHandler)
	e.GET("/login", handlers.LoginHandler)
	e.GET("/dashboard", handlers.DashboardHandler, mw.IsAuthenticated)

	// Auth routes
	auth := e.Group("auth")
	{
		discord := auth.Group("discord")
		{
			discord.GET("/", handlers.DiscordAuthHandler)
			discord.GET("/callback", handlers.DiscordAuthCallbackHandler)
		}
	}

	// API routes
	api := e.Group("/api", mw.IsAuthenticated)
	{
		api.GET("/reset-token", handlers.ResetTokenHandler)
		api.GET("/gallery", handlers.GalleryHandler)

		domains := api.Group("/domains")
		{
			domains.GET("/", handlers.ListDomainsHandler)
			domains.POST("/:name", handlers.CreateDomainHandler)
			//domains.DELETE("/:name". handlers.DeleteDomainHandler)
		}

		hosts := api.Group("/hosts")
		{
			hosts.GET("/", handlers.ListHostsHandler)
			hosts.POST("/:name", handlers.CreateHostHandler)
			hosts.DELETE("/:name", handlers.DeleteHostHandler)
		}
	}

}
