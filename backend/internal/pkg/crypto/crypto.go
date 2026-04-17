package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost   = 12
	apiKeyLength = 32
	apiKeyPrefix = "cm-"
)

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword verifies a password against its hash.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateAPIKey 生成一个新的 API Key，返回完整密钥、显示前缀和 SHA-256 哈希值。
func GenerateAPIKey() (fullKey, prefix, hash string, err error) {
	randomBytes := make([]byte, apiKeyLength)
	if _, err = rand.Read(randomBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	hexStr := hex.EncodeToString(randomBytes)
	fullKey = apiKeyPrefix + hexStr
	prefix = fullKey[:len(apiKeyPrefix)+8]
	hash = HashAPIKey(fullKey)

	return fullKey, prefix, hash, nil
}

// HashAPIKey computes SHA-256 hash of an API key.
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// GenerateRandomString generates a random hex string of specified length.
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
