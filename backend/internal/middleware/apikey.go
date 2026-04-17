package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// APIKeyInfo represents cached API key information in Redis.
type APIKeyInfo struct {
	DepartmentID *int64 `json:"department_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	UserID       int64  `json:"user_id"`
	KeyID        int64  `json:"key_id"`
	KeyStatus    int16  `json:"key_status"`
	UserStatus   int16  `json:"user_status"`
}

// API Key 认证中间件上下文键。
const (
	CtxKeyAPIKeyID    = "api_key_id"
	CtxKeyAPIKeyInfo  = "api_key_info"
	CtxKeyLLMProtocol = "llm_protocol"
)

// SetLLMProtocol sets the LLM protocol format for proper error responses.
func SetLLMProtocol(protocol string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(CtxKeyLLMProtocol, protocol)
		c.Next()
	}
}

// sendProtocolError sends error response in the appropriate protocol format.
func sendProtocolError(c *gin.Context, httpStatus int, msg string) {
	protocol, _ := c.Get(CtxKeyLLMProtocol)
	switch protocol {
	case "anthropic":
		c.JSON(httpStatus, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    anthropicErrorType(httpStatus),
				"message": msg,
			},
		})
	case "openai":
		c.JSON(httpStatus, gin.H{
			"error": gin.H{
				"message": msg,
				"type":    openaiErrorType(httpStatus),
				"param":   nil,
				"code":    nil,
			},
		})
	default:
		c.JSON(httpStatus, gin.H{
			"code":    httpStatus,
			"message": msg,
			"data":    nil,
		})
	}
}

func anthropicErrorType(status int) string {
	switch {
	case status == http.StatusUnauthorized:
		return "authentication_error"
	case status == http.StatusForbidden:
		return "permission_error"
	case status == http.StatusNotFound:
		return "not_found_error"
	case status == http.StatusTooManyRequests:
		return "rate_limit_error"
	case status >= 500: //nolint:mnd // intentional constant.
		return "api_error"
	default:
		return "invalid_request_error"
	}
}

func openaiErrorType(status int) string {
	switch {
	case status == http.StatusUnauthorized:
		return "invalid_api_key"
	case status == http.StatusForbidden:
		return "insufficient_quota"
	case status == http.StatusTooManyRequests:
		return "rate_limit_exceeded"
	case status >= 500: //nolint:mnd // intentional constant.
		return "server_error"
	default:
		return "invalid_request_error"
	}
}

// APIKeyAuth validates API keys from Authorization or x-api-key headers.
func APIKeyAuth(db *gorm.DB, rdb *redis.Client, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Detect self-loop: reject if request comes from CodeMind's own LLM client
		if c.GetHeader("X-CodeMind-Proxy") == "1" {
			logger.Error("request loop detected: LLM backend base_url may point to CodeMind itself",
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
			sendProtocolError(c, http.StatusBadGateway, "LLM backend misconfigured: base_url cannot point to CodeMind itself")
			c.Abort()
			return
		}

		apiKey := extractAPIKey(c)

		if apiKey == "" || !strings.HasPrefix(apiKey, "cm-") {
			if logger != nil {
				logger.Warn("invalid API key format",
					zap.String("path", c.Request.URL.Path),
					zap.Bool("is_empty", apiKey == ""),
					zap.Bool("has_cm_prefix", strings.HasPrefix(apiKey, "cm-")),
				)
			}
			sendProtocolError(c, errcode.ErrAPIKeyInvalid.HTTP, errcode.ErrAPIKeyInvalid.Message)
			c.Abort()
			return
		}

		keyHash := crypto.HashAPIKey(apiKey)

		if logger != nil {
			logger.Debug("API key hash computed",
				zap.String("key_hash_prefix", keyHash[:16]+"..."),
			)
		}

		info, err := getAPIKeyInfo(c.Request.Context(), db, rdb, keyHash)
		if err != nil {
			if logger != nil {
				logger.Warn("API key lookup failed",
					zap.String("key_hash_prefix", keyHash[:16]+"..."),
					zap.Error(err),
				)
			}
			sendProtocolError(c, errcode.ErrAPIKeyInvalid.HTTP, errcode.ErrAPIKeyInvalid.Message)
			c.Abort()
			return
		}

		if info.KeyStatus != 1 {
			sendProtocolError(c, errcode.ErrAPIKeyDisabled.HTTP, errcode.ErrAPIKeyDisabled.Message)
			c.Abort()
			return
		}

		if info.UserStatus != 1 {
			sendProtocolError(c, errcode.ErrAccountDisabled.HTTP, errcode.ErrAccountDisabled.Message)
			c.Abort()
			return
		}

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

// extractAPIKey extracts API key from headers (x-api-key takes precedence).
func extractAPIKey(c *gin.Context) string {
	if key := c.GetHeader("x-api-key"); key != "" {
		return key
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2) //nolint:mnd // intentional constant.
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	return ""
}

// getAPIKeyInfo retrieves API key info, with Redis cache.
func getAPIKeyInfo(ctx context.Context, db *gorm.DB, rdb *redis.Client, keyHash string) (*APIKeyInfo, error) {
	cacheKey := fmt.Sprintf("codemind:apikey:%s", keyHash)

	cached, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var info APIKeyInfo
		if json.Unmarshal([]byte(cached), &info) == nil {
			return &info, nil
		}
	}

	var result struct {
		DepartmentID *int64
		ExpiresAt    *time.Time
		Username     string
		Role         string
		KeyID        int64
		UserID       int64
		KeyStatus    int16
		UserStatus   int16
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
		return nil, fmt.Errorf("API key lookup failed: %w", err)
	}

	if result.ExpiresAt != nil && time.Now().After(*result.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
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

	if data, err := json.Marshal(info); err == nil {
		rdb.Set(ctx, cacheKey, string(data), 5*time.Minute) //nolint:mnd // intentional constant.
	}

	return info, nil
}
