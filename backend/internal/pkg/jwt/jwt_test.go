package jwt

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testJWTSecret is the JWT secret for unit tests (min 32 chars, required by NewManager).
const testJWTSecret = "test-secret-key-for-unit-testing-minimum-32-chars"

// setupTestManager creates a JWT Manager and miniredis for testing.
func setupTestManager(t *testing.T, expireHours int) (*Manager, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	manager, err := NewManager(testJWTSecret, expireHours, rdb)
	require.NoError(t, err)
	return manager, mr
}

func TestNewManager(t *testing.T) {
	t.Run("valid secret key creation succeeds", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		manager, err := NewManager(testJWTSecret, 24, rdb)

		require.NoError(t, err)
		require.NotNil(t, manager)
		assert.Equal(t, []byte(testJWTSecret), manager.secret)
		assert.Equal(t, 24, manager.expireHours)
		assert.Equal(t, rdb, manager.rdb)
	})

	t.Run("short secret key rejected", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		manager, err := NewManager("my-secret", 24, rdb)

		assert.Error(t, err)
		assert.Nil(t, manager)
	})
}

func TestGenerateToken(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	tests := []struct {
		deptID   *int64
		name     string
		username string
		role     string
		userID   int64
	}{
		{
			name:     "generate token with department ID",
			userID:   1,
			username: "admin",
			role:     "admin",
			deptID:   func() *int64 { v := int64(100); return &v }(),
		},
		{
			name:     "generate token without department ID",
			userID:   2,
			username: "user",
			role:     "user",
			deptID:   nil,
		},
		{
			name:     "generate token with nil department ID",
			userID:   3,
			username: "test",
			role:     "viewer",
			deptID:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, expiresAt, err := manager.GenerateToken(tt.userID, tt.username, tt.role, tt.deptID)

			require.NoError(t, err)
			assert.NotEmpty(t, tokenString)
			assert.True(t, expiresAt.After(time.Now()))

			// Verify token format
			parts := strings.Split(tokenString, ".")
			assert.Len(t, parts, 3, "JWT should have 3 parts")
		})
	}
}

func TestParseToken_Success(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	deptID := int64(100)
	tokenString, _, err := manager.GenerateToken(1, "admin", "admin", &deptID)
	require.NoError(t, err)

	claims, err := manager.ParseToken(tokenString)

	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	assert.Equal(t, "admin", claims.Role)
	assert.NotNil(t, claims.DepartmentID)
	assert.Equal(t, deptID, *claims.DepartmentID)
	assert.NotEmpty(t, claims.ID)
	assert.Equal(t, "codemind", claims.Issuer)
}

func TestParseToken_Expired(t *testing.T) {
	// Create manager with -1 hour expiration
	manager, mr := setupTestManager(t, -1)
	defer mr.Close()

	deptID := int64(100)
	tokenString, _, err := manager.GenerateToken(1, "admin", "admin", &deptID)
	require.NoError(t, err)

	claims, err := manager.ParseToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestParseToken_InvalidSignature(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	deptID := int64(100)
	tokenString, _, err := manager.GenerateToken(1, "admin", "admin", &deptID)
	require.NoError(t, err)

	// Create another manager with a different secret to parse
	wrongManager, wrongMr := setupTestManager(t, 24)
	defer wrongMr.Close()
	// Change secret to a different value
	wrongManager.secret = []byte("wrong-secret")

	claims, err := wrongManager.ParseToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestParseToken_Malformed(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	tests := []struct {
		name        string
		tokenString string
		wantErr     string
	}{
		{
			name:        "empty string",
			tokenString: "",
			wantErr:     "failed to parse token",
		},
		{
			name:        "invalid format",
			tokenString: "not-a-valid-token",
			wantErr:     "failed to parse token",
		},
		{
			name:        "missing parts",
			tokenString: "header.payload",
			wantErr:     "failed to parse token",
		},
		{
			name:        "base64 decode failure",
			tokenString: "invalid.base64.signature",
			wantErr:     "failed to parse token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ParseToken(tt.tokenString)

			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestParseToken_Blacklisted(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	// Generate token
	deptID := int64(100)
	tokenString, expiresAt, err := manager.GenerateToken(1, "admin", "admin", &deptID)
	require.NoError(t, err)

	// Parse once to get jti
	claims, err := manager.ParseToken(tokenString)
	require.NoError(t, err)
	jti := claims.ID

	// Add token to blacklist
	err = manager.Blacklist(context.Background(), jti, expiresAt)
	require.NoError(t, err)

	// Parsing again should fail
	claims, err = manager.ParseToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token has been revoked")
}

func TestBlacklist(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	tests := []struct {
		expiration time.Time
		name       string
		jti        string
		wantErr    bool
		wantBlack  bool
	}{
		{
			name:       "blacklist successfully",
			jti:        "test-jti-1",
			expiration: time.Now().Add(1 * time.Hour),
			wantErr:    false,
			wantBlack:  true,
		},
		{
			name:       "expired token does not need blacklisting",
			jti:        "test-jti-2",
			expiration: time.Now().Add(-1 * time.Hour),
			wantErr:    false,
			wantBlack:  false,
		},
		{
			name:       "zero remaining TTL does not need blacklisting",
			jti:        "test-jti-3",
			expiration: time.Now(),
			wantErr:    false,
			wantBlack:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Blacklist(context.Background(), tt.jti, tt.expiration)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check if blacklisted
			isBlacklisted := manager.IsBlacklisted(context.Background(), tt.jti)
			assert.Equal(t, tt.wantBlack, isBlacklisted)
		})
	}
}

func TestBlacklist_TTL(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	jti := "test-jti-ttl"
	expiration := time.Now().Add(30 * time.Minute)

	err := manager.Blacklist(context.Background(), jti, expiration)
	require.NoError(t, err)

	// Check TTL - miniredis TTL returns nanoseconds
	key := "codemind:jwt:blacklist:" + jti
	ttlNs := mr.TTL(key)
	ttlSec := ttlNs / 1e9
	// TTL should be close to 30 minutes (allowing some tolerance)
	assert.True(t, ttlSec > 29*60 && ttlSec <= 30*60, "TTL should be set correctly, got %d seconds", ttlSec)
}

func TestIsBlacklisted(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	tests := []struct {
		name     string
		setup    func()
		jti      string
		expected bool
	}{
		{
			name: "token not in blacklist",
			setup: func() {
				// No setup needed
			},
			jti:      "not-blacklisted",
			expected: false,
		},
		{
			name: "token in blacklist",
			setup: func() {
				key := "codemind:jwt:blacklist:blacklisted-token"
				mr.Set(key, "1")
				mr.SetTTL(key, time.Hour)
			},
			jti:      "blacklisted-token",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := manager.IsBlacklisted(context.Background(), tt.jti)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBlacklisted_RedisError(t *testing.T) {
	manager, mr := setupTestManager(t, 24)

	// Close Redis connection to simulate error
	mr.Close()

	// Fail-closed on Redis error, deny access
	result := manager.IsBlacklisted(context.Background(), "any-token")
	assert.True(t, result)
}

func TestIntegration_FullLifecycle(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	ctx := context.Background()
	deptID := int64(100)

	// 1. Generate Token
	tokenString, expiresAt, err := manager.GenerateToken(1, "admin", "admin", &deptID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// 2. Parse Token
	claims, err := manager.ParseToken(tokenString)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	jti := claims.ID
	assert.NotEmpty(t, jti)

	// 3. Confirm Token is not blacklisted
	assert.False(t, manager.IsBlacklisted(ctx, jti))

	// 4. Add Token to blacklist
	err = manager.Blacklist(ctx, jti, expiresAt)
	require.NoError(t, err)

	// 5. Confirm Token is blacklisted
	assert.True(t, manager.IsBlacklisted(ctx, jti))

	// 6. Parsing again should fail
	_, err = manager.ParseToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token has been revoked")
}

func TestClaims_WithDifferentSigningMethods(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	// Create a token with wrong signing algorithm
	claims := Claims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "test-id",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "codemind",
		},
	}

	// Use RS256 instead of HS256
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// Without a private key, we can only construct an invalid token
	tokenString, _ := token.SigningString()

	_, err := manager.ParseToken(tokenString + ".invalid")
	assert.Error(t, err)
}

func BenchmarkGenerateToken(b *testing.B) {
	mr := miniredis.RunT(&testing.T{})
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	manager, err := NewManager(testJWTSecret, 24, rdb)
	if err != nil {
		b.Fatal(err)
	}
	deptID := int64(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := manager.GenerateToken(1, "admin", "admin", &deptID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseToken(b *testing.B) {
	mr := miniredis.RunT(&testing.T{})
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	manager, err := NewManager(testJWTSecret, 24, rdb)
	if err != nil {
		b.Fatal(err)
	}
	deptID := int64(100)

	tokenString, _, err := manager.GenerateToken(1, "admin", "admin", &deptID)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ParseToken(tokenString)
		if err != nil {
			b.Fatal(err)
		}
	}
}
