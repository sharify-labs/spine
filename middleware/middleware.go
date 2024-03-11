package middleware

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth/gothic"
	"net/http"
	"os"
)

func Setup(e *echo.Echo) {
	// Init Gothic for oAuth2
	sessStore := sessions.NewCookieStore([]byte("secret")) // TODO: Secure this
	gothic.Store = sessStore

	// Init middleware
	e.Use(
		mw.Recover(),
		// TODO: Favicon
		mw.LoggerWithConfig(mw.LoggerConfig{
			Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
			Output: os.Stdout,
		}),
		session.Middleware(sessStore),
	)
}

// IsAuthenticated is a middleware that checks if the user is logged in.
func IsAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)
		if sess.Values["discord_username"] == nil {
			// If the user is not authenticated, redirect.
			return c.Redirect(http.StatusFound, "/login")
		}
		// If the session exists and is valid, proceed with the request.
		return next(c)
	}
}

//
//func RequiresAuth(c echo.Context) error {
//	sess, err := database.GetSession(c)
//	if err != nil {
//		return c.NoContent(http.StatusInternalServerError)
//	}
//	if sess.Get("user_id") == nil {
//		return c.Redirect("/")
//	}
//	return c.Next()
//}
