package clients

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sharify-labs/spine/config"
	"io"
	"net/http"
	"net/url"
)

var HTTP = &httpClient{}

type httpClient struct {
	client *http.Client
}

func (c *httpClient) Connect() {
	c.client = http.DefaultClient
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *httpClient) ForwardToZephyr(ctx echo.Context, userToken string) error {
	zephyrURL := &url.URL{
		Scheme:   "https",
		Host:     config.ZephyrURL,
		Path:     ctx.Path(),
		RawQuery: ctx.QueryString(),
	}

	req, err := http.NewRequest(ctx.Request().Method, zephyrURL.String(), ctx.Request().Body)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req.Header.Set("User-Agent", config.UserAgent+" "+ctx.Request().UserAgent())
	req.Header.Set(config.HeaderJWTAuth, userToken)
	req.Header.Set(config.HeaderSpineKey, config.Str("ZEPHYR_ADMIN_KEY"))
	req.Header.Set(echo.HeaderContentType, ctx.Request().Header.Get(echo.HeaderContentType))
	for name, values := range ctx.Request().Header {
		for _, val := range values {
			fmt.Printf("%s: %s", name, val)
		}
	}
	resp, err := c.Do(req)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSONBlob(resp.StatusCode, body)
}
