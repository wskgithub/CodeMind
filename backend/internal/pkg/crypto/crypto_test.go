package crypto

import (
	"strings"
	"testing"
)

// TestHashPassword tests password hashing.
func TestHashPassword(t *testing.T) {
	password := "Test@12345"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("password hashing failed: %v", err)
	}

	if hash == "" {
		t.Fatal("hash result should not be empty")
	}

	// Hash should not equal the original text
	if hash == password {
		t.Fatal("hash should not equal the original text")
	}
}

// TestCheckPassword tests password verification.
func TestCheckPassword(t *testing.T) {
	password := "Test@12345"
	hash, _ := HashPassword(password)

	// Correct password
	if !CheckPassword(password, hash) {
		t.Error("correct password should pass verification")
	}

	// Wrong password
	if CheckPassword("WrongPassword", hash) {
		t.Error("wrong password should not pass verification")
	}
}

// TestGenerateAPIKey tests API Key generation.
func TestGenerateAPIKey(t *testing.T) {
	fullKey, prefix, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("failed to generate API Key: %v", err)
	}

	// Verify complete key prefix format
	if !strings.HasPrefix(fullKey, "cm-") {
		t.Errorf("key should start with 'cm-', got: %s", fullKey)
	}

	// Verify key length: cm- + 64 hex = 67 characters
	if len(fullKey) != 67 {
		t.Errorf("key length should be 67, got: %d", len(fullKey))
	}

	// Verify prefix format: cm- followed by first 8 chars
	if !strings.HasPrefix(prefix, "cm-") {
		t.Errorf("prefix format is incorrect: %s", prefix)
	}

	// Verify hash is not empty
	if hash == "" {
		t.Fatal("hash result should not be empty")
	}

	// Re-hashing the same key should produce the same result
	rehash := HashAPIKey(fullKey)
	if rehash != hash {
		t.Error("same key should produce the same hash")
	}
}

// TestHashAPIKey tests API Key hashing.
func TestHashAPIKey(t *testing.T) {
	key1 := "cm-abcdef1234567890abcdef12345678"
	key2 := "cm-abcdef1234567890abcdef12345679"

	hash1 := HashAPIKey(key1)
	hash2 := HashAPIKey(key2)

	// Different keys should produce different hashes
	if hash1 == hash2 {
		t.Error("different keys should produce different hashes")
	}

	// Same key should produce same hash
	if HashAPIKey(key1) != hash1 {
		t.Error("same key should produce the same hash")
	}
}

// TestGenerateAPIKeyUniqueness tests API Key uniqueness.
func TestGenerateAPIKeyUniqueness(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, _, _, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("failed to generate API Key: %v", err)
		}
		if keys[key] {
			t.Fatalf("generated duplicate key: %s", key)
		}
		keys[key] = true
	}
}
