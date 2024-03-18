package router

import (
	"github.com/labstack/echo/v4"
	"github.com/sharify-labs/spine/handlers"
	mw "github.com/sharify-labs/spine/middleware"
)

// Setup initializes all routes:
//
// Root:
// GET     / -> handlers.Root
// GET     /login -> handlers.Login
//
// Auth:
// GET     /auth/discord -> handlers.DiscordAuth
// GET     /auth/discord/callback -> handlers.DiscordAuthCallback
//
// Protected:
// GET     /dashboard -> handlers.DisplayDashboard
// GET     /api/reset-token -> handlers.ResetToken
// GET     /api/gallery -> handlers.DisplayGallery
//
// GET     /api/hosts -> handlers.ListHosts
// POST    /api/hosts -> handlers.CreateHost
// DELETE  /api/hosts -> handlers.DeleteHost
func Setup(e *echo.Echo) {
	// Root routes
	e.GET("", handlers.Root)
	e.GET("/login", handlers.Login)
	e.GET("/dashboard", handlers.DisplayDashboard, mw.IsAuthenticated)

	// Auth routes
	auth := e.Group("/auth")
	{
		discord := auth.Group("/discord")
		{
			discord.GET("", handlers.DiscordAuth)
			discord.GET("/callback", handlers.DiscordAuthCallback)
		}
	}

	// API routes
	api := e.Group("/api", mw.IsAuthenticated)
	{
		api.GET("/reset-token", handlers.ResetToken)
		api.GET("/gallery", handlers.DisplayGallery)

		domains := api.Group("/domains")
		{
			domains.GET("", handlers.ListAvailableDomains)
		}

		hosts := api.Group("/hosts")
		{
			hosts.GET("", handlers.ListHosts)
			hosts.POST("", handlers.CreateHost)
			hosts.DELETE("/:name", handlers.DeleteHost)
		}
	}

}
