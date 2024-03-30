package middleware

import (
	"embed"
	"encoding/gob"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth/gothic"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/models"
	"net/http"
	"os"
	"time"
)

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
		mw.Recover(),
		mw.BodyLimit("100M"),
		mw.LoggerWithConfig(mw.LoggerConfig{
			Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
			Output: os.Stdout,
		}),
		mw.CORSWithConfig(mw.CORSConfig{
			AllowOrigins: config.List("ALLOW_ORIGINS"),
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

// RequireSession is a middleware that checks if the user is logged in.
func RequireSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			c.Logger().Errorf("unable to get session: %v", err)
			return c.Redirect(http.StatusFound, "/login")
		}
		if sess.Values["auth_user"] != nil {
			return next(c) // Session is valid, proceed with the request.
		}
		return c.Redirect(http.StatusFound, "/login")
	}
}
