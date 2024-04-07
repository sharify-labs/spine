package services

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/database"
	"gorm.io/gorm/clause"
	"time"
)

const zephyrTokenPrefix string = "sfy"

type ZephyrToken struct {
	Value string
}

// GenerateJWT creates a new JWT for the user.
// This token is stored in the user's Cookies so that it can be used
// to authenticate with Zephyr when uploading directly from the web panel.
func GenerateJWT(userID string) (string, error) {
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(config.SessionMaxAge)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Subject:   userID,
		Issuer:    "spine",
	}).SignedString(config.JWTPrivateKey)
	if err != nil {
		return "", err
	}
	return tokenStr, err
}

// NewZephyrToken generates a new upload token and stores it in the database.
//
// We don't need to salt API Tokens:
// https://security.stackexchange.com/questions/209936/do-i-need-to-use-salt-with-api-key-hashing
//
// Raw Tokens are in the format:
// `sfy_<16-chars>_<32-chars>`
// `sfy_<id>_<key>`
// Example: sfy_3c9c0fe69b72b2c1_734c5c796877fb00f2fc31d024c62f12302367f08338dc35113b42eef7be7fd3
//
// Notes:
//   - Token ID and Key are hex-encoded for user
//   - Token ID and Key Hash are base64-RawURLEncoded in database.
//   - Token ID is generated the first time a user receives a token.
//   - When tokens get refreshed, the ID stays the same. Only the 'key' changes.
//   - The key is hashed and stored. The id acts as a 'username'.
func NewZephyrToken(userID string) (*ZephyrToken, error) {
	// Get user from database
	var user database.User
	tx := database.DB().Begin()
	if tx.Error != nil {
		tx.Rollback()
		return nil, tx.Error
	}
	if err := tx.Clauses(clause.Locking{
		Strength: clause.LockingStrengthUpdate,
	}).Preload("Token").Where(&database.User{
		ID: userID,
	}).First(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Check if user has existing token
	var err error
	var tokenID []byte
	if user.Token != nil {
		tokenID, err = base64.RawURLEncoding.DecodeString(user.Token.ID)
	} else {
		tokenID, err = GenerateRandomBytes(8)
		if err != nil {
			return nil, err
		}
	}

	// Generate token key
	key, err := GenerateRandomBytes(32)
	if err != nil {
		return nil, err
	}
	hash, err := Hash(key)
	if err != nil {
		return nil, err
	}

	// Store in database
	user.Token = &database.Token{
		ID:     base64.RawURLEncoding.EncodeToString(tokenID),
		Hash:   base64.RawURLEncoding.EncodeToString(hash),
		UserID: userID,
	}
	if err = tx.Clauses(clause.Locking{
		Strength: clause.LockingStrengthUpdate,
	}).Save(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Combine prefix, id, and key, seperated by underscores and return to user.
	// Example: sfy_3c9c0fe69b72b2c1_734c5c796877fb00f2fc31d024c62f12302367f08338dc35113b42eef7be7fd3
	return &ZephyrToken{
		Value: zephyrTokenPrefix + "_" + hex.EncodeToString(tokenID) + "_" + hex.EncodeToString(key),
	}, nil
}

// GenerateRandomBytes generates a byte array with given length
// randomly and securely using CSPRNG in the crypto/rand package
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err // rand.Read should only fail if the system's secure RNG fails.
	}
	return bytes, nil
}

// Hash hashes a byte array with SHA-512 and returns the hash.
func Hash(data []byte) ([]byte, error) {
	hasher := sha512.New()
	if _, err := hasher.Write(data); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}
