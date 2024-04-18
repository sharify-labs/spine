package clients

import (
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	echolog "github.com/labstack/gommon/log"
	"github.com/sharify-labs/spine/config"
)

var Sentry = &sentryClient{}

type sentryClient struct{}

func (*sentryClient) Connect() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: config.Get[string]("SENTRY_DSN"),
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	}); err != nil {
		echolog.Warnf("Sentry initialization failed: %v", err)
	}
}

func (*sentryClient) CaptureErr(c echo.Context, err error) {
	c.Logger().Error(err)
	if hub := sentryecho.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
	}
}
