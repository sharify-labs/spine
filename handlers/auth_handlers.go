package handlers

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/models"
	"github.com/sharify-labs/spine/services"
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
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// Ensure user is entered into Database
	user, err := database.GetOrCreateUser(discordUser)
	if err != nil {
		c.Logger().Errorf("failed to get/create user (database): %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// Generate temp key (used for uploading to Zephyr directly from panel)
	zephyrJWT, err := services.GenerateJWT(user.ID)
	if err != nil {
		c.Logger().Errorf("failed to generate JWT: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	c.Logger().Debugf("Generated new JWT: %s", zephyrJWT)
	authUser := models.AuthorizedUser{
		ID:        user.ID,
		ZephyrJWT: zephyrJWT,
	}
	authUser.Discord.Username = discordUser.Name
	authUser.Discord.Email = discordUser.Email

	// Store user's details in session
	sess, err := session.Get("session", c)
	if err != nil {
		c.Logger().Errorf("failed getting session in Discord callback: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	sess.Values["auth_user"] = authUser

	// Save session
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		c.Logger().Errorf("failed saving session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
