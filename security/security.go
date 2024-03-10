package security

import (
	"crypto/rand"
	"crypto/sha512"
	"github.com/google/uuid"
)

type ZephyrKey struct {
	Key  string
	Salt []byte
	Hash []byte
}

// NewZephyrKey generates a new upload key. Does NOT store it in database.
func NewZephyrKey() *ZephyrKey {
	key := uuid.NewString()
	salt, err := generateSalt(16)
	if err != nil {
		return nil
	}
	return &ZephyrKey{
		Key:  key,
		Salt: salt,
		Hash: hashKey(key, salt),
	}
}

// generateSalt generates n bytes randomly and securely
// using CSPRNG in the crypto/rand package
func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	return salt, err
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
