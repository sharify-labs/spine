package config

import (
	"encoding/base64"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	HeaderJWTAuth  string = "Authorization" // Used for Zephyr Auth
	HeaderSpineKey string = "X-Spine-Key"   // Used to verify HeaderJWTAuth is coming from spine (ZEPHYR_ADMIN_KEY)
	HostDefault    string = "sharify.me"
	ZephyrURL      string = "xericl.dev"
	UserAgent      string = "sharify-labs/spine"
)
const (
	SessionMaxAge = time.Hour * 24 * 7
)

// Str reads in a string variable from environment.
// Note: Panics if value is empty.
func Str(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing config value for " + key)
	}
	return value
}

// DecodedB64 reads in a base64-encoded string from environment and decodes it.
// Additionally, it validates that the result is the expected length.
// Note: Panics if value is empty.
func DecodedB64(key string, length int) []byte {
	value, err := base64.StdEncoding.DecodeString(Str(key))
	if err != nil {
		panic(err)
	}
	if len(value) != length {
		panic("base64 string for " + key + " is not expected length " + strconv.Itoa(length))
	}
	return value
}

// Int reads in a string variable from environment and converts it to an integer.
// Note: Panics if value is empty.
func Int(key string) int {
	value, err := strconv.Atoi(Str(key))
	if err != nil {
		panic("invalid integer value for " + key)
	}
	return value
}

// Bool reads in a string variable from environment and converts it to a boolean.
// Note: Panics if value is empty.
func Bool(key string) bool {
	value, err := strconv.ParseBool(Str(key))
	if err != nil {
		panic("invalid boolean value for " + key)
	}
	return value
}

// List reads in a string variable from environment and splits it where there are commas to create a list.
// Note: Panics if value is empty.
func List(key string) []string {
	value := Str(key)
	if value == "" {
		panic("missing config value for " + key)
	}
	return strings.Split(value, ",")
}
