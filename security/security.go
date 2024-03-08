package security

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/posty/spine/config"
	"github.com/posty/spine/models"
	"strings"
)

const (
	ApiKeyHeader string = "X-API-Key"
)

// ValidateUploadKey validates the UserID and Key of the request
func ValidateUploadKey(c *fiber.Ctx) error {
	key := c.Get(ApiKeyHeader, "")
	unauthorized := c.Status(fiber.StatusUnauthorized).SendString("invalid key")

	if strings.TrimSpace(key) == "" {
		return unauthorized
	}

	// Check environment for API key
	storedKey := models.Key{
		Hash: config.GetStr("API_KEY"),
		Salt: config.GetStr("SECRET"),
	}

	// Check if key is valid
	isValid := validateKey(key, &storedKey)

	if isValid {
		log.Debug("Validated API key")
		return c.Next() // user provided valid API key
	}

	return unauthorized
}

func validateKey(key string, storedKey *models.Key) bool {
	var inputHash []byte

	storedHash, err := base64.StdEncoding.DecodeString(storedKey.Hash)
	if err != nil {
		log.Error("Error decoding stored hash")
		return false
	}

	storedSalt, err := base64.StdEncoding.DecodeString(storedKey.Salt)
	if err != nil {
		log.Error("Error decoding stored salt")
		return false
	}

	// Hash key with salt from database
	inputHash = hashKey(key, storedSalt)

	// Compare hashes
	if subtle.ConstantTimeCompare(inputHash[:], storedHash[:]) == 1 {
		return true
	}
	log.Info("Hashes don't match")

	return false
}

// HashKey computes the SHA-512 hash of the key using the salt
func hashKey(key string, salt []byte) []byte {
	keyBytes := []byte(key)
	hasher := sha512.New()

	// Append salt to key
	keyBytes = append(keyBytes, salt...)
	hasher.Write(keyBytes)

	// Generate SHA512 hash
	return hasher.Sum(nil)
}
