package handlers

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/posty/spine/database"
	"github.com/posty/spine/models"
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

	var err error
	var authUser goth.User
	authUser, err = gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Store user's Discord details in session
	sess, err := session.Get("session", c)
	sess.Values["discord_username"] = authUser.Name
	sess.Values["discord_email"] = authUser.Email

	// Ensure user is entered into Database
	var u *models.User
	u, err = database.GetOrCreateUser(authUser.Email)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Store UserID in session
	sess.Values["user_id"] = u.ID

	// Save session
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
