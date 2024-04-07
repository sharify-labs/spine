package clients

import (
	"fmt"
	goccy "github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/database"
	"io"
	"net/http"
	"net/url"
	"time"
)

var HTTP = &httpClient{}

type httpClient struct {
	client *http.Client
}

func (c *httpClient) Connect() {
	c.client = http.DefaultClient
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "sharify-labs/spine")
	return c.client.Do(req)
}

// GetOrFetchAvailableDomains gets list of available domains from cache or fetches them from our GitHub repo.
func (c *httpClient) GetOrFetchAvailableDomains() (map[string]interface{}, error) {
	const cacheKey string = "cache:available_domains"

	domains := make(map[string]interface{})
	database.GetFromCache(cacheKey, &domains)
	fmt.Printf("\nRetreived %d domains from cache", len(domains))
	// TODO: do this in zephyr for list uploads endpoint
	if len(domains) == 0 {
		const domainsURL = "https://gist.githubusercontent.com/xEricL/91d1a37fa70f0964a31c700f39416118/raw/d8dd06da809202bac5ff46f6e8ec0d34a0e484a6/domains.json"
		req, err := http.NewRequest(http.MethodGet, domainsURL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, err
		}
		var githubResp struct {
			Domains map[string]interface{} `json:"domains"`
		}
		if err = goccy.NewDecoder(resp.Body).Decode(&githubResp); err != nil {
			return nil, err
		}
		domains = githubResp.Domains
		database.AddToCache(cacheKey, domains, 12*time.Hour)
	}

	return domains, nil
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
		Sentry.CaptureErr(ctx, err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req.Header.Set("User-Agent", config.UserAgent+" "+ctx.Request().UserAgent())
	req.Header.Set(config.HeaderJWTAuth, userToken)
	req.Header.Set(config.HeaderSpineKey, config.Str("ZEPHYR_ADMIN_KEY"))
	req.Header.Set(echo.HeaderContentType, ctx.Request().Header.Get(echo.HeaderContentType))
	for name, values := range ctx.Request().Header {
		for _, val := range values {
			ctx.Logger().Debugf("Header %s: %s", name, val)
		}
	}
	resp, err := c.Do(req)
	if err != nil {
		Sentry.CaptureErr(ctx, err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Sentry.CaptureErr(ctx, err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSONBlob(resp.StatusCode, body)
}
