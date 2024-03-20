package middleware

import (
	"embed"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth/gothic"
	"github.com/sharify-labs/spine/config"
	"net/http"
	"os"
	"time"
)

func Setup(e *echo.Echo, assets embed.FS) {
	// Init Gothic for oAuth2
	sessStore := sessions.NewCookieStore(
		config.GetDecodeB64("SESSION_AUTH_KEY_64"),
		config.GetDecodeB64("SESSION_ENC_KEY_32"),
	)
	gothic.Store = sessStore

	// Init middleware
	e.Use(
		mw.Recover(),
		mw.BodyLimit("100M"),
		mw.LoggerWithConfig(mw.LoggerConfig{
			Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
			Output: os.Stdout,
		}),
		mw.CORSWithConfig(mw.CORSConfig{
			AllowOrigins: config.GetList("ALLOW_ORIGINS"),
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
}

// IsAuthenticated is a middleware that checks if the user is logged in.
func IsAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil || sess.Values["discord_username"] == nil {
			// If the user is not authenticated, redirect.
			return c.Redirect(http.StatusFound, "/login")
		}

		// If the session exists and is valid, proceed with the request.
		return next(c)
	}
}
