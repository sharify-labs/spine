package middleware

import (
	"embed"
	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/posty/spine/database"
	"time"
)

func Setup(a *fiber.App, assets embed.FS) {
	a.Use(
		// Sentry.io middleware
		fibersentry.New(fibersentry.Config{
			Timeout: 3 * time.Second, WaitForDelivery: true},
		),
		// TODO: Favicon
		//Logger middleware
		logger.New(),
	)
}

func RequiresAuth(c *fiber.Ctx) error {
	sess, err := database.GetSession(c)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if sess.Get("user_id") == nil {
		return c.Redirect("/")
	}
	return c.Next()
}
