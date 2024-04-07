package config

import (
	"crypto/ecdsa"
	"encoding/base64"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"os"
	"strconv"
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

var (
	JWTPrivateKey *ecdsa.PrivateKey
)

// Setup reads .env and initializes values that aren't const but also shouldn't be read in from env more than once.
func Setup() {
	var err error
	if err = godotenv.Load(); err != nil {
		panic(err)
	}
	var jwtPem []byte
	if jwtPem, err = base64.StdEncoding.DecodeString(Get[string]("JWT_PRIVATE_KEY")); err != nil {
		panic("failed to decode JWT_PRIVATE_KEY")
	}
	if JWTPrivateKey, err = jwt.ParseECPrivateKeyFromPEM(jwtPem); err != nil {
		panic("failed to ParseECPrivateKeyFromPEM(JWT_PRIVATE_KEY)")
	}
}

// Get reads in a value from environment variable and returns its value as specified type.
func Get[T any](key string) T {
	value := os.Getenv(key)
	if value == "" {
		panic("missing config value for " + key)
	}
	var result any
	switch any(new(T)).(type) {
	case *string:
		result = value
	case *int:
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			panic("invalid integer value for " + key)
		}
		result = valueInt
	default:
		panic("unsupported type")
	}
	return result.(T)
}

// DecodedB64 reads in a base64-encoded string from environment and decodes it.
// Additionally, it validates that the result is the expected length.
// Note: Panics if value is empty.
func DecodedB64(key string, length int) []byte {
	value, err := base64.StdEncoding.DecodeString(Get[string](key))
	if err != nil {
		panic(err)
	}
	if len(value) != length {
		panic("base64 string for " + key + " is not expected length " + strconv.Itoa(length))
	}
	return value
}
