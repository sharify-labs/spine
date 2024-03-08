package middleware

import (
	"embed"
	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"time"
)

func Setup(a *fiber.App, assets embed.FS) {
	a.Use(
		// Sentry.io middleware
		fibersentry.New(fibersentry.Config{
			Timeout: 3 * time.Second, WaitForDelivery: true},
		),
		//Logger middleware
		logger.New(),
	)
}
