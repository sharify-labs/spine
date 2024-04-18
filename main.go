package main

import (
	"context"
	"embed"
	"errors"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/router"
)

//go:embed assets/*
var assets embed.FS
var version string

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	config.Setup()

	e := echo.New()
	e.Logger.SetLevel(log.Lvl(config.Get[int]("LOG_LEVEL")))
	e.IPExtractor = echo.ExtractIPFromXFFHeader() // internal IPs trusted by default
	e.Renderer = &Template{
		templates: template.Must(template.ParseFS(assets, "assets/templates/*.html")),
	}

	goth.UseProviders(
		discord.New(
			config.Get[string]("DISCORD_CLIENT_ID"),
			config.Get[string]("DISCORD_CLIENT_SECRET"),
			config.Get[string]("DISCORD_CALLBACK_URL"),
			"identify", "email",
		),
	)

	clients.Setup()
	database.Setup()
	router.Setup(e, assets)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start app
	go func() {
		e.Logger.Infof("Started Spine %s", version)
		if err := e.Start(":" + config.Get[string]("PORT")); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatalf("shutting down server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully close the server with a timeout of 30 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	} else {
		e.Logger.Info("server closed gracefully")
	}
}
