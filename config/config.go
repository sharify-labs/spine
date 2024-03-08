package config

import (
	"embed"
	goccy "github.com/goccy/go-json"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"
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

func GetTrustedProxies(assets embed.FS) []string {
	const path string = "assets/cloudflare_ips.json"
	var out []string

	file, err := fs.ReadFile(assets, path)
	if err != nil {
		log.Fatalf("Unable to read %s", path)
	}

	err = goccy.Unmarshal(file, &out)
	if err != nil {
		log.Fatalf("Unable to unmarshal %s", path)
	}

	// Append any IPs added to .env
	proxies := strings.Split(os.Getenv("TRUSTED_PROXIES_LIST"), ",")
	if len(proxies) >= 1 {
		for _, p := range proxies {
			out = append(out, p)
		}
	}
	log.Println(out)
	return out
}
