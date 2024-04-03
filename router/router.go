package router

import (
	"github.com/labstack/echo/v4"
	h "github.com/sharify-labs/spine/handlers"
	mw "github.com/sharify-labs/spine/middleware"
)

// Setup initializes all routes:
//
// Root:
// - GET     /       -> handlers.Root
// - GET     /login  -> handlers.Login
//
// Auth:
// - GET     /auth/discord           -> handlers.DiscordAuth
// - GET     /auth/discord/callback  -> handlers.DiscordAuthCallback
//
// Protected:
// - GET     /dashboard       		-> handlers.DisplayDashboard
// - GET     /api/v1/reset-token 	-> handlers.ResetToken
// - GET	   /api/v1/config/:type -> handlers.ProvideConfig  // :type must be files/pastes/redirects
// - GET	   /api/v1/domains      -> handlers.ListAvailableDomains
// - GET     /api/v1/hosts        	-> handlers.ListHosts
// - POST    /api/v1/hosts        	-> handlers.CreateHost
// - DELETE  /api/v1/hosts/:name  	-> handlers.DeleteHost
//
// Zephyr Routes:
//   - These routes forward the body and query parameters directly to Zephyr.
//   - All paths must exactly match the paths in Zephyr (see clients.HTTP ForwardToZephyr method)
//   - They are protected just like API routes.
//
// - GET /api/v1/uploads 		-> handlers.ZephyrProxy
// - POST /api/v1/uploads 		-> handlers.ZephyrProxy
// - DELETE /api/v1/uploads 	-> handlers.ZephyrProxy
func Setup(e *echo.Echo) {
	// Root routes
	e.GET("", h.Root)
	e.GET("/login", h.Login)

	// Auth routes
	auth := e.Group("/auth")
	{
		auth.GET("/discord", h.DiscordAuth)
		auth.GET("/discord/callback", h.DiscordAuthCallback)
	}

	// Protected routes
	e.GET("/dashboard", h.DisplayDashboard, mw.RequireSession)
	api := e.Group("/api", mw.RequireSession)
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/reset-token", h.ResetToken)
			v1.GET("/config/:type", h.ProvideConfig) // TODO: Make this 1 endpoint that downloads a zip with all configs
			v1.GET("/domains", h.ListAvailableDomains)

			v1.GET("/hosts", h.ListHosts)
			v1.POST("/hosts", h.CreateHost)
			v1.DELETE("/hosts/:name", h.DeleteHost)

			v1.GET("/uploads", h.ZephyrProxy)
			v1.POST("/uploads", h.ZephyrProxy)
			v1.DELETE("/uploads", h.ZephyrProxy)
		}
	}
}
