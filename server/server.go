package server

import (
	"embed"
	goccy "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"github.com/posty/spine/config"
	"github.com/posty/spine/database"
	"log"
)

//go:embed assets/*
var assets embed.FS

func Run(version string) {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize Sentry
	//lib.StartSentry(version)

	// Set Log Level for Fiber
	fiberlog.SetLevel(fiberlog.Level(config.GetInt("LOG_LEVEL")))

	// Create app
	app := fiber.New(fiber.Config{
		AppName:                 "Zephyr v" + version,
		ProxyHeader:             fiber.HeaderXForwardedFor,
		BodyLimit:               100 * 1024 * 1024, // 100mb
		JSONEncoder:             goccy.Marshal,
		JSONDecoder:             goccy.Unmarshal,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          config.GetTrustedProxies(assets),
	})

	// Connect to Database
	database.ConnectDB()

	// Setup middleware
	middleware.Setup(app, assets)

	// Setup router
	router.Setup(app)

	// Send console message to alert Pterodactyl
	log.Println("Started Zephyr v" + version)

	// Start app
	log.Fatal(app.Listen(":" + config.GetStr("PORT")))
}
