package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// APIKeyInfo 缓存在 Redis 中的 API Key 信息
type APIKeyInfo struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	DepartmentID *int64 `json:"department_id"`
	KeyID        int64  `json:"key_id"`
	KeyStatus    int16  `json:"key_status"`
	UserStatus   int16  `json:"user_status"`
}

// 上下文键常量（LLM 代理请求专用）
const (
	CtxKeyAPIKeyID   = "api_key_id"
	CtxKeyAPIKeyInfo = "api_key_info"
)

// APIKeyAuth API Key 认证中间件
// 支持两种认证方式：
//   - Authorization: Bearer cm-xxx （OpenAI 兼容格式）
//   - x-api-key: cm-xxx （Anthropic 原生格式）
func APIKeyAuth(db *gorm.DB, rdb *redis.Client, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 自环检测：若请求来自 CodeMind 自身的 LLM Client 转发，立即拒绝
		if c.GetHeader("X-CodeMind-Proxy") == "1" {
			logger.Error("检测到请求自环：LLM 后端 base_url 可能指向了 CodeMind 自身，请检查 LLM 节点配置",
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(502, gin.H{
				"error": gin.H{
					"message": "LLM 后端配置错误：base_url 不能指向 CodeMind 自身（检测到请求自环）",
					"type":    "configuration_error",
					"code":    "502",
				},
			})
			c.Abort()
			return
		}

		// 从 Header 提取 API Key（兼容 OpenAI 和 Anthropic 两种认证方式）
		apiKey := extractAPIKey(c)
		
		// 调试日志：记录收到的请求头（生产环境应移除）
		if logger != nil {
			authHeader := c.GetHeader("Authorization")
			xApiKey := c.GetHeader("x-api-key")
			logger.Debug("API Key 认证请求",
				zap.String("path", c.Request.URL.Path),
				zap.String("authorization_header", authHeader),
				zap.String("x_api_key_header", xApiKey),
				zap.String("extracted_key_prefix", func() string {
					if len(apiKey) > 10 {
						return apiKey[:10] + "..."
					}
					return apiKey
				}()),
			)
		}
		
		if apiKey == "" || !strings.HasPrefix(apiKey, "cm-") {
			if logger != nil {
				logger.Warn("API Key 格式无效",
					zap.String("path", c.Request.URL.Path),
					zap.Bool("is_empty", apiKey == ""),
					zap.Bool("has_cm_prefix", strings.HasPrefix(apiKey, "cm-")),
				)
			}
			response.Error(c, errcode.ErrAPIKeyInvalid)
			c.Abort()
			return
		}

		// 计算 Key 的 SHA-256 哈希
		keyHash := crypto.HashAPIKey(apiKey)
		
		if logger != nil {
			logger.Debug("API Key 哈希计算",
				zap.String("key_hash_prefix", keyHash[:16]+"..."),
			)
		}

		// 查询 Key 信息（优先从 Redis 缓存获取）
		info, err := getAPIKeyInfo(c.Request.Context(), db, rdb, keyHash)
		if err != nil {
			if logger != nil {
				logger.Warn("API Key 查询失败",
					zap.String("key_hash_prefix", keyHash[:16]+"..."),
					zap.Error(err),
				)
			}
			response.Error(c, errcode.ErrAPIKeyInvalid)
			c.Abort()
			return
		}

		// 验证 Key 状态
		if info.KeyStatus != 1 {
			response.Error(c, errcode.ErrAPIKeyDisabled)
			c.Abort()
			return
		}

		// 验证用户状态
		if info.UserStatus != 1 {
			response.Error(c, errcode.ErrAccountDisabled)
			c.Abort()
			return
		}

		// 将信息注入上下文
		c.Set(CtxKeyUserID, info.UserID)
		c.Set(CtxKeyUsername, info.Username)
		c.Set(CtxKeyRole, info.Role)
		c.Set(CtxKeyAPIKeyID, info.KeyID)
		c.Set(CtxKeyAPIKeyInfo, info)
		if info.DepartmentID != nil {
			c.Set(CtxKeyDepartmentID, *info.DepartmentID)
		}

		c.Next()
	}
}

// extractAPIKey 从请求头提取 API Key
// 优先级: x-api-key > Authorization: Bearer
func extractAPIKey(c *gin.Context) string {
	// 1. 尝试 Anthropic 风格: x-api-key 头
	if key := c.GetHeader("x-api-key"); key != "" {
		return key
	}

	// 2. 尝试 OpenAI 风格: Authorization: Bearer xxx
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	return ""
}

// getAPIKeyInfo 获取 API Key 关联信息，优先查 Redis 缓存
func getAPIKeyInfo(ctx context.Context, db *gorm.DB, rdb *redis.Client, keyHash string) (*APIKeyInfo, error) {
	cacheKey := fmt.Sprintf("codemind:apikey:%s", keyHash)

	// 1. 尝试从 Redis 获取
	cached, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var info APIKeyInfo
		if json.Unmarshal([]byte(cached), &info) == nil {
			return &info, nil
		}
	}

	// 2. 从数据库查询
	// 使用原生 SQL 查询以避免 GORM 自动推断排序字段
	var result struct {
		KeyID        int64
		KeyStatus    int16
		UserID       int64
		Username     string
		Role         string
		DepartmentID *int64
		UserStatus   int16
		ExpiresAt    *time.Time
	}

	query := `
		SELECT
			api_keys.id as key_id,
			api_keys.status as key_status,
			api_keys.expires_at,
			users.id as user_id,
			users.username,
			users.role,
			users.department_id,
			users.status as user_status
		FROM api_keys
		JOIN users ON users.id = api_keys.user_id AND users.deleted_at IS NULL
		WHERE api_keys.key_hash = $1
		LIMIT 1
	`

	err = db.Raw(query, keyHash).Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("API Key 查询失败: %w", err)
	}

	// 检查 Key 是否过期
	if result.ExpiresAt != nil && time.Now().After(*result.ExpiresAt) {
		return nil, fmt.Errorf("API Key 已过期")
	}

	info := &APIKeyInfo{
		UserID:       result.UserID,
		Username:     result.Username,
		Role:         result.Role,
		DepartmentID: result.DepartmentID,
		KeyID:        result.KeyID,
		KeyStatus:    result.KeyStatus,
		UserStatus:   result.UserStatus,
	}

	// 3. 写入 Redis 缓存（TTL 5 分钟）
	if data, err := json.Marshal(info); err == nil {
		rdb.Set(ctx, cacheKey, string(data), 5*time.Minute)
	}

	return info, nil
}
