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

// testJWTSecret 单元测试用 JWT 密钥（至少 32 字符，满足 NewManager 校验）.
const testJWTSecret = "test-secret-key-for-unit-testing-minimum-32-chars"

// setupTestManager 创建测试用的 JWT Manager 和 miniredis.
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
	t.Run("合法密钥创建成功", func(t *testing.T) {
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

	t.Run("短密钥被拒绝", func(t *testing.T) {
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
			name:     "生成带部门ID的Token",
			userID:   1,
			username: "admin",
			role:     "admin",
			deptID:   func() *int64 { v := int64(100); return &v }(),
		},
		{
			name:     "生成不带部门ID的Token",
			userID:   2,
			username: "user",
			role:     "user",
			deptID:   nil,
		},
		{
			name:     "生成空部门ID的Token",
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

			// 验证 token 格式
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
	// 创建过期时间为 -1 小时的管理器
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

	// 创建另一个使用不同 secret 的管理器来解析
	wrongManager, wrongMr := setupTestManager(t, 24)
	defer wrongMr.Close()
	// 修改 secret 为不同的值
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
			name:        "空字符串",
			tokenString: "",
			wantErr:     "failed to parse token",
		},
		{
			name:        "非法格式",
			tokenString: "not-a-valid-token",
			wantErr:     "failed to parse token",
		},
		{
			name:        "缺少部分",
			tokenString: "header.payload",
			wantErr:     "failed to parse token",
		},
		{
			name:        "base64解码失败",
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

	// 生成 token
	deptID := int64(100)
	tokenString, expiresAt, err := manager.GenerateToken(1, "admin", "admin", &deptID)
	require.NoError(t, err)

	// 先解析一次获取 jti
	claims, err := manager.ParseToken(tokenString)
	require.NoError(t, err)
	jti := claims.ID

	// 将 token 加入黑名单
	err = manager.Blacklist(context.Background(), jti, expiresAt)
	require.NoError(t, err)

	// 再次解析应该失败
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
			name:       "加入黑名单成功",
			jti:        "test-jti-1",
			expiration: time.Now().Add(1 * time.Hour),
			wantErr:    false,
			wantBlack:  true,
		},
		{
			name:       "已过期不需要加入黑名单",
			jti:        "test-jti-2",
			expiration: time.Now().Add(-1 * time.Hour),
			wantErr:    false,
			wantBlack:  false,
		},
		{
			name:       "剩余有效期为0不需要加入黑名单",
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

			// 检查是否在黑名单中
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

	// 检查 TTL - miniredis 的 TTL 返回纳秒
	key := "codemind:jwt:blacklist:" + jti
	ttlNs := mr.TTL(key)
	ttlSec := ttlNs / 1e9
	// TTL 应该接近 30 分钟（允许一些误差）
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
			name: "Token不在黑名单中",
			setup: func() {
				// 不做任何设置
			},
			jti:      "not-blacklisted",
			expected: false,
		},
		{
			name: "Token在黑名单中",
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

	// 关闭 Redis 连接以模拟错误
	mr.Close()

	// Redis 错误时 fail-closed，拒绝放行
	result := manager.IsBlacklisted(context.Background(), "any-token")
	assert.True(t, result)
}

func TestIntegration_FullLifecycle(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	ctx := context.Background()
	deptID := int64(100)

	// 1. 生成 Token
	tokenString, expiresAt, err := manager.GenerateToken(1, "admin", "admin", &deptID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// 2. 解析 Token
	claims, err := manager.ParseToken(tokenString)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	jti := claims.ID
	assert.NotEmpty(t, jti)

	// 3. 确认 Token 不在黑名单
	assert.False(t, manager.IsBlacklisted(ctx, jti))

	// 4. 将 Token 加入黑名单
	err = manager.Blacklist(ctx, jti, expiresAt)
	require.NoError(t, err)

	// 5. 确认 Token 在黑名单中
	assert.True(t, manager.IsBlacklisted(ctx, jti))

	// 6. 再次解析应该失败
	_, err = manager.ParseToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token has been revoked")
}

func TestClaims_WithDifferentSigningMethods(t *testing.T) {
	manager, mr := setupTestManager(t, 24)
	defer mr.Close()

	// 创建一个使用错误签名算法的 token
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

	// 使用 RS256 而不是 HS256
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// 由于没有私钥，我们只能构造一个无效的 token
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
