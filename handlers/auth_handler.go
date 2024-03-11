package handlers

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"github.com/posty/spine/database"
	"net/http"
)

func DiscordAuthHandler(c echo.Context) error {
	q := c.Request().URL.Query()
	q.Add("provider", "discord")
	c.Request().URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(c.Response().Writer, c.Request())
	return nil
}

func DiscordAuthCallbackHandler(c echo.Context) error {
	q := c.Request().URL.Query()
	q.Add("provider", "discord")
	c.Request().URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Store user's Discord details in session
	sess, err := session.Get("session", c)
	sess.Values["discord_username"] = user.Name
	sess.Values["discord_email"] = user.Email
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Ensure user is entered into Database
	_, err = database.GetOrCreateUser(user.Email)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusFound, "/dashboard")
}

func LoginHandler(c echo.Context) error {
	return c.HTML(http.StatusOK, `<a href="/auth/discord">Login with Discord</a>`)
}
