package config

import (
	"embed"
	"encoding/base64"
	"fmt"
	goccy "github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	UserHeader  string = "X-Upload-User"
	HostHeader  string = "X-Upload-Host"
	DefaultHost string = "sharify.me"
)

func GetStr(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing config value for " + key)
	}
	return value
}

// GetDecodeB64 reads a string from environmental variables and decodes it with base64.
// It is used for reading secrets and includes a length arg for safety to ensure secret is desired length.
func GetDecodeB64(key string, length int) []byte {
	value, err := base64.StdEncoding.DecodeString(GetStr(key))
	if err != nil {
		panic(err)
	}
	if len(value) != length {
		panic("base64 string for " + key + " is not expected length " + strconv.Itoa(length))
	}
	return value
}

func GetInt(key string) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		panic("missing config value for " + key)
	}
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		panic("invalid integer value for " + key)
	}
	return valueInt
}

func GetBool(key string) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		panic("missing config value for " + key)
	}
	valueBool, err := strconv.ParseBool(valueStr)
	if err != nil {
		panic("invalid integer boolean for " + key)
	}
	return valueBool
}

func GetList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing config value for " + key)
	}
	return strings.Split(value, ",")
}

func GetTrustedProxyRanges(assets embed.FS) []echo.TrustOption {
	const path string = "assets/cloudflare_ips.json"
	var out []echo.TrustOption
	var ipRanges []string

	file, err := fs.ReadFile(assets, path)
	if err != nil {
		panic("unable to read " + path)
	}

	err = goccy.Unmarshal(file, &ipRanges)
	if err != nil {
		panic("unable to unmarshal " + path)
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
