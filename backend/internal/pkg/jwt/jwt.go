package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Claims 自定义 JWT 声明
type Claims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	DepartmentID *int64 `json:"department_id,omitempty"`
	jwt.RegisteredClaims
}

// Manager JWT 管理器
type Manager struct {
	secret      []byte
	expireHours int
	rdb         *redis.Client
}

// NewManager 创建 JWT 管理器
func NewManager(secret string, expireHours int, rdb *redis.Client) *Manager {
	return &Manager{
		secret:      []byte(secret),
		expireHours: expireHours,
		rdb:         rdb,
	}
}

// GenerateToken 生成 JWT Token
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
		return "", time.Time{}, fmt.Errorf("签发 Token 失败: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ParseToken 解析并验证 JWT Token
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Token 解析失败: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Token 无效")
	}

	// 检查黑名单
	if m.IsBlacklisted(context.Background(), claims.ID) {
		return nil, fmt.Errorf("Token 已被注销")
	}

	return claims, nil
}

// Blacklist 将 Token 加入黑名单（登出时调用）
func (m *Manager) Blacklist(ctx context.Context, jti string, expiration time.Time) error {
	// 计算剩余有效期作为 TTL
	ttl := time.Until(expiration)
	if ttl <= 0 {
		return nil // 已过期，无需加入黑名单
	}

	key := fmt.Sprintf("codemind:jwt:blacklist:%s", jti)
	return m.rdb.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted 检查 Token 是否在黑名单中
func (m *Manager) IsBlacklisted(ctx context.Context, jti string) bool {
	key := fmt.Sprintf("codemind:jwt:blacklist:%s", jti)
	result, err := m.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false // Redis 错误时不阻塞，降级放行
	}
	return result > 0
}
