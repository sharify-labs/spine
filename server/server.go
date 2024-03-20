package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/middleware"
	"github.com/sharify-labs/spine/router"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func Start(assets embed.FS, version string) {
	// Load .env
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	// Initialize Sentry
	//lib.StartSentry(version)

	// Create app
	e := echo.New()

	// Configure trusted proxies
	e.IPExtractor = echo.ExtractIPFromXFFHeader(config.GetTrustedProxyRanges(assets)...)

	// Setup HTML Template rendering
	e.Renderer = &Template{
		templates: template.Must(template.ParseFS(assets, "assets/templates/*.html")),
	}

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
	middleware.Setup(e, assets)

	// Setup router
	router.Setup(e)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start app
	go func() {
		fmt.Println("Started Spine v" + version)
		if err := e.Start(":" + config.GetStr("PORT")); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatalf("shutting down server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully close the server with a timeout of 30 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
