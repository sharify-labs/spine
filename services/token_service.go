package services

import (
	"crypto/rand"
	"crypto/sha512"
	"github.com/sharify-labs/spine/database"
)

type ZephyrToken struct {
	Value string
	salt  []byte
	hash  []byte
}

// NewZephyrToken generates a new upload token. Does NOT store it in database.
func NewZephyrToken() (*ZephyrToken, error) {
	value, err := GenerateRandomStringHex(32)
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
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

// GenerateRandomStringHex generates a random hex string with given length
// randomly and securely using CSPRNG in the crypto/rand package
func GenerateRandomStringHex(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err // rand.Read should only fail if the system's secure RNG fails.
	}
	return hex.EncodeToString(bytes), nil
}
