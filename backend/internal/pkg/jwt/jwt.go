package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const minSecretLength = 32

// Claims holds custom JWT claims
type Claims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	DepartmentID *int64 `json:"department_id,omitempty"`
	jwt.RegisteredClaims
}

// Manager handles JWT operations
type Manager struct {
	secret      []byte
	expireHours int
	rdb         *redis.Client
}

// NewManager creates a JWT manager, returns error if secret is too short
func NewManager(secret string, expireHours int, rdb *redis.Client) (*Manager, error) {
	if len(secret) < minSecretLength {
		return nil, fmt.Errorf("JWT secret too short: need at least %d chars, got %d", minSecretLength, len(secret))
	}
	return &Manager{
		secret:      []byte(secret),
		expireHours: expireHours,
		rdb:         rdb,
	}, nil
}

// GenerateToken creates a new JWT token
func (m *Manager) GenerateToken(userID int64, username, role string, deptID *int64) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(m.expireHours) * time.Hour)
	jti := uuid.New().String()

	claims := Claims{
		UserID:       userID,
		Username:     username,
		Role:         role,
		DepartmentID: deptID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "codemind",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ParseToken validates and parses a JWT token
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Strictly enforce HS256
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if m.IsBlacklisted(context.Background(), claims.ID) {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}

// Blacklist adds a token to the blacklist
func (m *Manager) Blacklist(ctx context.Context, jti string, expiration time.Time) error {
	ttl := time.Until(expiration)
	if ttl <= 0 {
		return nil
	}

	key := fmt.Sprintf("codemind:jwt:blacklist:%s", jti)
	return m.rdb.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted checks if a token is blacklisted
// Fail-closed: returns true on Redis errors to prevent revoked tokens from being used
func (m *Manager) IsBlacklisted(ctx context.Context, jti string) bool {
	key := fmt.Sprintf("codemind:jwt:blacklist:%s", jti)
	result, err := m.rdb.Exists(ctx, key).Result()
	if err != nil {
		return true
	}
	return result > 0
}
