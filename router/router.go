package router

import (
	"embed"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth/gothic"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/config"
	h "github.com/sharify-labs/spine/handlers"
	"github.com/sharify-labs/spine/models"
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
// - GET	 /api/v1/config/:type 	-> handlers.ProvideConfig  // :type must be files/pastes/redirects
// - GET	 /api/v1/domains      	-> handlers.ListAvailableDomains
// - GET     /api/v1/hosts        	-> handlers.ListHosts
// - POST    /api/v1/hosts        	-> handlers.CreateHost
// - DELETE  /api/v1/hosts/:name  	-> handlers.DeleteHost
//
// Zephyr Routes:
//
//   - These routes forward the body and query parameters directly to Zephyr.
//
//   - All paths must exactly match the paths in Zephyr (see clients.HTTP ForwardToZephyr method)
//
//   - They are protected just like API routes.
//
//   - GET /api/v1/uploads
//
//   - POST /api/v1/uploads
//
//   - DELETE /api/v1/uploads
func Setup(e *echo.Echo, assets embed.FS) {
	// Init Gothic for oAuth2
	sessStore := sessions.NewCookieStore(
		config.DecodedB64("SESSION_AUTH_KEY_64", 64),
		config.DecodedB64("SESSION_ENC_KEY_32", 32),
	)
	sessStore.MaxAge(int(config.SessionMaxAge.Seconds()))
	gothic.Store = sessStore
	gob.Register(models.AuthorizedUser{})

	// Init middleware
	e.Use(
		mw.Secure(),
		mw.Recover(),
		mw.BodyLimit("100M"),
		mw.LoggerWithConfig(mw.LoggerConfig{
			Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
			Output: os.Stdout,
		}),
		mw.CORSWithConfig(mw.CORSConfig{
			AllowOrigins: strings.Split(config.Get[string]("ALLOW_ORIGINS"), ","),
		}),
		mw.StaticWithConfig(mw.StaticConfig{
			Root:       "assets/static",
			Filesystem: http.FS(assets),
		}),
		sentryecho.New(sentryecho.Options{
			Timeout: 3 * time.Second,
			Repanic: true,
		}),
		session.Middleware(sessStore),
	)

	e.GET("", h.Root)
	e.GET("/login", h.Login)

	auth := e.Group("/auth")
	{
		auth.GET("/discord", h.DiscordAuth)
		auth.GET("/discord/callback", h.DiscordAuthCallback)
	}

	// Protected routes
	e.GET("/dashboard", h.DisplayDashboard, requireSession)
	api := e.Group("/api", requireSession)
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

// requireSession is a middleware that checks if the user is logged in.
func requireSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			clients.Sentry.CaptureErr(c, fmt.Errorf("unable to get session: %w", err))
			return c.Redirect(http.StatusFound, "/login")
		}
		if user, ok := sess.Values["auth_user"].(models.AuthorizedUser); ok {
			c.Set("user", user)
			return next(c)
		}
		return c.Redirect(http.StatusFound, "/login")
	}
}
