package config

import (
	"embed"
	"fmt"
	goccy "github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"io/fs"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	UserHeader string = "X-Upload-User"
	HostHeader string = "X-Upload-Host"
)

func GetStr(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Missing config value for %s", key)
	}
	return value
}

func GetInt(key string) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		log.Fatalf("Missing config value for %s", key)
	}
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("Invalid integer value for %s: %s", key, valueStr)
	}
	return valueInt
}

func GetBool(key string) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		log.Fatalf("Missing config value for %s", key)
	}
	valueBool, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Fatalf("Invalid boolean value for %s: %s", key, valueStr)
	}
	return valueBool
}

func GetList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Missing config value for %s", key)
	}
	return strings.Split(value, ",")
}

func GetTrustedProxyRanges(assets embed.FS) []echo.TrustOption {
	const path string = "assets/cloudflare_ips.json"
	var out []echo.TrustOption
	var ipRanges []string

	file, err := fs.ReadFile(assets, path)
	if err != nil {
		log.Fatalf("Unable to read %s", path)
	}

	err = goccy.Unmarshal(file, &ipRanges)
	if err != nil {
		log.Fatalf("Unable to unmarshal %s", path)
	}

	var ipNet *net.IPNet
	for _, r := range ipRanges {
		_, ipNet, err = net.ParseCIDR(r)
		if err != nil {
			fmt.Printf("IP range %q could not be parsed: %v\n", r, err)
		} else {
			out = append(out, echo.TrustIPRange(ipNet))
		}
	}

	return out
}
