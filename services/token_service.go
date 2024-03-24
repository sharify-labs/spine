package services

import (
	"crypto/rand"
	"crypto/sha512"
	"github.com/google/uuid"
	"github.com/sharify-labs/spine/database"
)

type ZephyrToken struct {
	Value string
	salt  []byte
	hash  []byte
}

// NewZephyrToken generates a new upload token. Does NOT store it in database.
func NewZephyrToken() (*ZephyrToken, error) {
	value := uuid.NewString()
	salt, err := generateSalt(16)
	if err != nil {
		return nil, err
	}
	return &ZephyrToken{
		Value: value,
		salt:  salt,
		hash:  hashToken(value, salt),
	}, nil
}

func (zt *ZephyrToken) AssignToUser(userID string) error {
	return database.UpdateUserToken(userID, zt.hash, zt.salt)
}

// generateSalt generates n bytes randomly and securely
// using CSPRNG in the crypto/rand package
func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	return salt, err
}

// hashToken computes the SHA-512 hash of the token using the salt
func hashToken(value string, salt []byte) []byte {
	tokenBytes := []byte(value)
	hasher := sha512.New()

	// Append salt to token
	tokenBytes = append(tokenBytes, salt...)
	hasher.Write(tokenBytes)

	// Generate SHA512 hash
	return hasher.Sum(nil)
}
