package handlers

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/models"
	"net/http"
)

func DiscordAuth(c echo.Context) error {
	q := c.Request().URL.Query()
	q.Add("provider", "discord")
	c.Request().URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(c.Response().Writer, c.Request())
	return nil
}

func DiscordAuthCallback(c echo.Context) error {
	q := c.Request().URL.Query()
	q.Add("provider", "discord")
	c.Request().URL.RawQuery = q.Encode()

	discordUser, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	if err != nil {
		c.Logger().Errorf("failed to complete gothic user auth: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// Ensure user is entered into Database
	user, err := database.GetOrCreateUser(discordUser.Email)
	if err != nil {
		c.Logger().Errorf("failed to get/create user (database): %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	authUser := models.AuthorizedUser{ID: user.ID}
	authUser.Discord.Username = discordUser.Name
	authUser.Discord.Email = discordUser.Email

	// Store user's details in session
	sess, err := session.Get("session", c)
	sess.Values["auth_user"] = authUser

	// Save session
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		c.Logger().Errorf("failed saving session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
