package clients

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/sharify-labs/spine/config"
)

var Sentry = &sentryClient{}

type sentryClient struct{}

func (_ *sentryClient) Connect() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: config.GetStr("SENTRY_DSN"),
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v", err)
	}
}

func (_ *sentryClient) CaptureErr(c echo.Context, err error) {
	if hub := sentryecho.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
	}
}
