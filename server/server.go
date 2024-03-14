package server

import (
	"embed"
	"errors"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/posty/spine/clients"
	"github.com/posty/spine/config"
	"github.com/posty/spine/database"
	"github.com/posty/spine/middleware"
	"github.com/posty/spine/router"
	"log"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

//type Template struct {
//	templates *template.Template
//}

func Start() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	// Initialize Sentry
	//lib.StartSentry(version)

	// Set Log Level for Fiber
	//fiberlog.SetLevel(fiberlog.Level(config.GetInt("LOG_LEVEL")))
	//fiber.Config{
	//	ProxyHeader:             fiber.HeaderXForwardedFor,
	//	JSONEncoder:             goccy.Marshal,
	//	JSONDecoder:             goccy.Unmarshal,
	//	EnableTrustedProxyCheck: true,
	//	TrustedProxies:          config.GetTrustedProxies(assets),
	//}

	// Create app
	e := echo.New()

	//// Setup HTML Template rendering
	//e.Renderer = &Template{
	//	templates: template.Must(template.ParseGlob("views/*.html")),
	//}

	// Setup Goth Auth Providers
	goth.UseProviders(
		discord.New(
			config.GetStr("DISCORD_CLIENT_ID"),
			config.GetStr("DISCORD_CLIENT_SECRET"),
			config.GetStr("DISCORD_CALLBACK_URL"),
			"identify", "email",
		),
	)

	// Setup clients
	clients.Setup()

	// Setup databases
	database.Setup()

	// Setup middleware
	middleware.Setup(e)

	// Setup router
	router.Setup(e)

	// Send console message to alert Pterodactyl
	log.Println("Started Spine")

	// Start app
	if err := e.Start(":" + config.GetStr("PORT")); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

//
//func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
//	return t.templates.ExecuteTemplate(w, name, data)
//}
