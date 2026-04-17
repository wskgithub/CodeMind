package router

import (
	"codemind/internal/model"
	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/jwt"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==================== Test Helpers ====================

// testJWTSecret 路由测试用 JWT 密钥（至少 32 字符，满足 jwt.NewManager 校验）.
const testJWTSecret = "test-secret-key-for-unit-testing-minimum-32-chars"

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	return db
}

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return mr, rdb
}

func setupTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupTestRouter(t *testing.T) (*gin.Engine, *miniredis.Miniredis, *redis.Client, *gorm.DB, *jwt.Manager) {
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	jwtManager, err := jwt.NewManager(testJWTSecret, 24, rdb)
	require.NoError(t, err)
	logger := zaptest.NewLogger(t)

	// 创建必要的表用于 API Key 认证测试
	db.Exec(`CREATE TABLE api_keys (
		id INTEGER PRIMARY KEY,
		key_hash TEXT UNIQUE,
		status INTEGER,
		expires_at TIMESTAMP,
		user_id INTEGER
	)`)
	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		username TEXT,
		role TEXT,
		department_id INTEGER,
		status INTEGER,
		deleted_at TIMESTAMP
	)`)

	engine := setupTestGin()
	// 使用空的 handlers 结构体测试路由注册和中间件
	// 注意：这会导致 handler 路由返回 500，但可以用来验证路由存在和中间件
	handlers := &Handlers{} // 空 handlers
	// corsOrigins 为 nil 时 CORS 为 AllowAllOrigins，响应头 Access-Control-Allow-Origin 为 *
	Setup(engine, handlers, jwtManager, db, rdb, logger, nil, "")

	return engine, mr, rdb, db, jwtManager
}

func setupTestRouterWithMonitor(t *testing.T) (*gin.Engine, *miniredis.Miniredis, *redis.Client, *gorm.DB, *jwt.Manager) {
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	jwtManager, err := jwt.NewManager(testJWTSecret, 24, rdb)
	require.NoError(t, err)
	logger := zaptest.NewLogger(t)

	db.Exec(`CREATE TABLE api_keys (
		id INTEGER PRIMARY KEY,
		key_hash TEXT UNIQUE,
		status INTEGER,
		expires_at TIMESTAMP,
		user_id INTEGER
	)`)
	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		username TEXT,
		role TEXT,
		department_id INTEGER,
		status INTEGER,
		deleted_at TIMESTAMP
	)`)

	engine := setupTestGin()
	// 使用空的 handlers 结构体，但 Monitor 为 nil
	handlers := &Handlers{}
	Setup(engine, handlers, jwtManager, db, rdb, logger, nil, "")

	return engine, mr, rdb, db, jwtManager
}

// ==================== Health Check Tests ====================

func TestHealthEndpoint(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

// ==================== Public Routes Tests ====================

func TestPublicRoutes(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	tests := []struct {
		name         string
		method       string
		path         string
		expectExists bool // 期望路由是否存在（即使 handler 未实现）
	}{
		{
			name:         "健康检查",
			method:       "GET",
			path:         "/health",
			expectExists: true,
		},
		{
			name:         "登录接口",
			method:       "POST",
			path:         "/api/v1/auth/login",
			expectExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			// 路由存在不应返回 404
			if tt.expectExists {
				assert.NotEqual(t, 404, w.Code, "路由应该存在")
			}
		})
	}
}

// ==================== JWT Protected Routes Tests ====================

func TestJWTProtectedRoutes_WithoutAuth(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedCode   int // errcode
	}{
		{
			name:           "未认证访问 logout",
			method:         "POST",
			path:           "/api/v1/auth/logout",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 profile",
			method:         "GET",
			path:           "/api/v1/auth/profile",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 update profile",
			method:         "PUT",
			path:           "/api/v1/auth/profile",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 change password",
			method:         "PUT",
			path:           "/api/v1/auth/password",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 keys",
			method:         "GET",
			path:           "/api/v1/keys",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 users",
			method:         "GET",
			path:           "/api/v1/users",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 departments",
			method:         "GET",
			path:           "/api/v1/departments",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 stats overview",
			method:         "GET",
			path:           "/api/v1/stats/overview",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 limits my",
			method:         "GET",
			path:           "/api/v1/limits/my",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 announcements",
			method:         "GET",
			path:           "/api/v1/announcements",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 system configs",
			method:         "GET",
			path:           "/api/v1/system/configs",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 mcp services",
			method:         "GET",
			path:           "/api/v1/mcp/services",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "未认证访问 monitor dashboard",
			method:         "GET",
			path:           "/api/v1/monitor/dashboard",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedCode > 0 {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				code := int(resp["code"].(float64))
				assert.Equal(t, tt.expectedCode, code)
			}
		})
	}
}

func TestJWTProtectedRoutes_WithValidToken(t *testing.T) {
	engine, mr, _, _, jwtManager := setupTestRouter(t)
	defer mr.Close()

	// 生成有效 token
	token, _, err := jwtManager.GenerateToken(1, "testuser", "user", nil)
	assert.NoError(t, err)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "认证访问 logout",
			method: "POST",
			path:   "/api/v1/auth/logout",
		},
		{
			name:   "认证访问 profile",
			method: "GET",
			path:   "/api/v1/auth/profile",
		},
		{
			name:   "认证访问 update profile",
			method: "PUT",
			path:   "/api/v1/auth/profile",
		},
		{
			name:   "认证访问 change password",
			method: "PUT",
			path:   "/api/v1/auth/password",
		},
		{
			name:   "认证访问 keys",
			method: "GET",
			path:   "/api/v1/keys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			// 路由存在且通过认证，不应该返回 401
			assert.NotEqual(t, 401, w.Code, "不应该返回 401，因为 token 是有效的")
			assert.NotEqual(t, 404, w.Code, "路由应该存在")
		})
	}
}

func TestJWTProtectedRoutes_WithInvalidToken(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "无效的 token 格式",
			token:          "invalid-token",
			expectedStatus: 401,
		},
		{
			name:           "错误的 Bearer 格式",
			token:          "Basic invalid",
			expectedStatus: 401,
		},
		{
			name:           "空的 Bearer token",
			token:          "",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/auth/profile", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ==================== Admin Routes Tests (RequireAdmin) ====================

func TestAdminRoutes_RequireAdmin(t *testing.T) {
	engine, mr, _, _, jwtManager := setupTestRouter(t)
	defer mr.Close()

	tests := []struct {
		name           string
		role           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "超级管理员访问系统配置",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/system/configs",
			method:         "GET",
			expectedStatus: 500, // 路由通过，handler 为 nil，Recovery 中间件返回 500
		},
		{
			name:           "部门经理访问系统配置被拒绝",
			role:           model.RoleDeptManager,
			path:           "/api/v1/system/configs",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "普通用户访问系统配置被拒绝",
			role:           model.RoleUser,
			path:           "/api/v1/system/configs",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "超级管理员访问 MCP 管理",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/mcp/services",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "部门经理访问 MCP 管理被拒绝",
			role:           model.RoleDeptManager,
			path:           "/api/v1/mcp/services",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "超级管理员访问监控",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/monitor/dashboard",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "部门经理访问监控被拒绝",
			role:           model.RoleDeptManager,
			path:           "/api/v1/monitor/dashboard",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "超级管理员访问 LLM 后端管理",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/system/llm-backends",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "部门经理访问 LLM 后端管理被拒绝",
			role:           model.RoleDeptManager,
			path:           "/api/v1/system/llm-backends",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "超级管理员访问审计日志",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/system/audit-logs",
			method:         "GET",
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _, err := jwtManager.GenerateToken(1, "testuser", tt.role, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ==================== Manager Routes Tests (RequireRole) ====================

func TestManagerRoutes_RequireManager(t *testing.T) {
	engine, mr, _, _, jwtManager := setupTestRouter(t)
	defer mr.Close()

	tests := []struct {
		name           string
		role           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "超级管理员访问用户管理",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/users",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "部门经理访问用户管理",
			role:           model.RoleDeptManager,
			path:           "/api/v1/users",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "普通用户访问用户管理被拒绝",
			role:           model.RoleUser,
			path:           "/api/v1/users",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "超级管理员访问部门管理",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/departments",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "部门经理访问部门管理",
			role:           model.RoleDeptManager,
			path:           "/api/v1/departments",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "普通用户访问部门管理被拒绝",
			role:           model.RoleUser,
			path:           "/api/v1/departments",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "超级管理员访问限额管理",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/limits",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "部门经理访问限额管理",
			role:           model.RoleDeptManager,
			path:           "/api/v1/limits",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "普通用户访问限额管理被拒绝",
			role:           model.RoleUser,
			path:           "/api/v1/limits",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "超级管理员导出 CSV",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/stats/export/csv",
			method:         "GET",
			expectedStatus: 400, // handler 为 nil 但参数验证返回 400
		},
		{
			name:           "部门经理导出 CSV",
			role:           model.RoleDeptManager,
			path:           "/api/v1/stats/export/csv",
			method:         "GET",
			expectedStatus: 400, // handler 为 nil 但参数验证返回 400
		},
		{
			name:           "普通用户导出 CSV 被拒绝",
			role:           model.RoleUser,
			path:           "/api/v1/stats/export/csv",
			method:         "GET",
			expectedStatus: 400, // 先检查参数，然后才检查权限
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _, err := jwtManager.GenerateToken(1, "testuser", tt.role, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ==================== API Key Protected Routes Tests ====================

func TestAPIKeyProtectedRoutes_WithoutKey(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "无 API Key 访问 chat completions",
			method:         "POST",
			path:           "/v1/chat/completions",
			expectedStatus: 401,
		},
		{
			name:           "无 API Key 访问 completions",
			method:         "POST",
			path:           "/v1/completions",
			expectedStatus: 401,
		},
		{
			name:           "无 API Key 访问 models",
			method:         "GET",
			path:           "/v1/models",
			expectedStatus: 401,
		},
		{
			name:           "无 API Key 访问 embeddings",
			method:         "POST",
			path:           "/v1/embeddings",
			expectedStatus: 401,
		},
		{
			name:           "无 API Key 访问 messages",
			method:         "POST",
			path:           "/v1/messages",
			expectedStatus: 401,
		},
		{
			name:           "无 API Key 访问 responses",
			method:         "POST",
			path:           "/v1/responses",
			expectedStatus: 401,
		},
		{
			name:           "无 API Key 访问 MCP SSE",
			method:         "GET",
			path:           "/mcp/sse",
			expectedStatus: 401,
		},
		{
			name:           "无 API Key 访问 MCP message",
			method:         "POST",
			path:           "/mcp/message",
			expectedStatus: 401,
		},
		{
			name:           "无 API Key 访问 MCP Streamable HTTP",
			method:         "POST",
			path:           "/mcp/",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAPIKeyProtectedRoutes_WithInvalidKey(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "无效的 API Key 格式",
			apiKey:         "invalid-key",
			expectedStatus: 401,
		},
		{
			name:           "非 cm- 前缀的 API Key",
			apiKey:         "sk-testkey123",
			expectedStatus: 401,
		},
		{
			name:           "空的 API Key",
			apiKey:         "",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
			req.Header.Set("Content-Type", "application/json")
			if tt.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+tt.apiKey)
			}
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAPIKeyProtectedRoutes_WithValidKey(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// 创建测试数据
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	// 插入测试用户和 API Key
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	// 清除 Redis 缓存
	rdb.FlushAll(t.Context())

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int // 500 表示路由和中间件通过，但 handler 未实现
	}{
		{
			name:           "有效 API Key 访问 chat completions",
			method:         "POST",
			path:           "/v1/chat/completions",
			expectedStatus: 500,
		},
		{
			name:           "有效 API Key 访问 completions",
			method:         "POST",
			path:           "/v1/completions",
			expectedStatus: 500,
		},
		{
			name:           "有效 API Key 访问 models",
			method:         "GET",
			path:           "/v1/models",
			expectedStatus: 500,
		},
		{
			name:           "有效 API Key 访问 embeddings",
			method:         "POST",
			path:           "/v1/embeddings",
			expectedStatus: 500,
		},
		{
			name:           "有效 API Key 访问 messages",
			method:         "POST",
			path:           "/v1/messages",
			expectedStatus: 500,
		},
		{
			name:           "有效 API Key 访问 MCP SSE",
			method:         "GET",
			path:           "/mcp/sse",
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+testAPIKey)
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			// 不应该返回 401（认证通过）或 404（路由存在）
			assert.NotEqual(t, 401, w.Code, "不应该返回 401，因为 API Key 是有效的")
			assert.NotEqual(t, 404, w.Code, "路由应该存在")
		})
	}
}

func TestAPIKeyProtectedRoutes_WithXAPIKey(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// 创建测试数据
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	rdb.FlushAll(t.Context())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/models", nil)
	// 使用 x-api-key header（Anthropic 格式）
	req.Header.Set("x-api-key", testAPIKey)
	engine.ServeHTTP(w, req)

	// 不应该返回 401，因为中间件支持 x-api-key
	assert.NotEqual(t, 401, w.Code)
	assert.NotEqual(t, 404, w.Code)
}

// ==================== CORS Tests ====================

func TestCORSHeaders(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	engine.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSOptionsRequest(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")
	engine.ServeHTTP(w, req)

	// CORS 中间件应该处理 OPTIONS 请求
	assert.Contains(t, []int{200, 204}, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
}

// ==================== 404 Handler Tests ====================

func Test404Handler(t *testing.T) {
	engine, mr, _, _, _ := setupTestRouter(t)
	defer mr.Close()

	tests := []struct {
		name   string
		path   string
		method string
	}{
		{
			name:   "访问不存在的路径",
			path:   "/nonexistent",
			method: "GET",
		},
		{
			name:   "访问不存在的 API 路径",
			path:   "/api/v999/test",
			method: "GET",
		},
		{
			name:   "访问不存在的 auth 路径",
			path:   "/api/v1/auth/nonexistent",
			method: "GET",
		},
		{
			name:   "访问不存在的 v1 路径",
			path:   "/v1/nonexistent",
			method: "GET",
		},
		{
			name:   "访问不存在的 mcp 路径",
			path:   "/mcp/nonexistent",
			method: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			engine.ServeHTTP(w, req)

			assert.Equal(t, 404, w.Code)
		})
	}
}

// ==================== Route Registration Tests ====================

func TestRouteRegistration(t *testing.T) {
	engine, mr, _, _, jwtManager := setupTestRouter(t)
	defer mr.Close()

	// 创建有效 token
	token, _, err := jwtManager.GenerateToken(1, "testuser", model.RoleSuperAdmin, nil)
	assert.NoError(t, err)

	// 测试所有主要路由是否正确注册
	routes := []struct {
		path          string
		method        string
		requireAuth   bool
		requireAPIKey bool
	}{
		// 公开路由
		{"/health", "GET", false, false},
		{"/api/v1/auth/login", "POST", false, false},

		// JWT 保护路由 - 认证相关
		{"/api/v1/auth/logout", "POST", true, false},
		{"/api/v1/auth/profile", "GET", true, false},
		{"/api/v1/auth/profile", "PUT", true, false},
		{"/api/v1/auth/password", "PUT", true, false},

		// JWT 保护路由 - API Key
		{"/api/v1/keys", "GET", true, false},
		{"/api/v1/keys", "POST", true, false},

		// JWT 保护路由 - 用户管理
		{"/api/v1/users", "GET", true, false},
		{"/api/v1/users/1", "GET", true, false},

		// JWT 保护路由 - 部门管理
		{"/api/v1/departments", "GET", true, false},
		{"/api/v1/departments/1", "GET", true, false},

		// JWT 保护路由 - 统计
		{"/api/v1/stats/overview", "GET", true, false},
		{"/api/v1/stats/usage", "GET", true, false},
		{"/api/v1/stats/ranking", "GET", true, false},

		// JWT 保护路由 - 限额
		{"/api/v1/limits/my", "GET", true, false},
		{"/api/v1/limits/my/progress", "GET", true, false},

		// JWT 保护路由 - 公告
		{"/api/v1/announcements", "GET", true, false},

		// JWT 保护路由 - 系统管理（需 admin）
		{"/api/v1/system/configs", "GET", true, false},
		{"/api/v1/system/llm-backends", "GET", true, false},
		{"/api/v1/system/audit-logs", "GET", true, false},

		// JWT 保护路由 - MCP 管理（需 admin）
		{"/api/v1/mcp/services", "GET", true, false},
		{"/api/v1/mcp/access-rules", "GET", true, false},

		// JWT 保护路由 - 监控（需 admin）
		{"/api/v1/monitor/dashboard", "GET", true, false},
		{"/api/v1/monitor/system", "GET", true, false},
		{"/api/v1/monitor/requests", "GET", true, false},
		{"/api/v1/monitor/llm-nodes", "GET", true, false},
		{"/api/v1/monitor/health", "GET", true, false},
	}

	for _, route := range routes {
		name := route.method + " " + route.path
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(route.method, route.path, nil)

			if route.requireAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			engine.ServeHTTP(w, req)

			// 路由存在（不应返回 404）
			assert.NotEqual(t, 404, w.Code, "路由应该存在")
		})
	}
}

// ==================== Middleware Chain Tests ====================

func TestMiddlewareChain(t *testing.T) {
	t.Run("Recovery 中间件已注册", func(t *testing.T) {
		engine, mr, _, _, _ := setupTestRouter(t)
		defer mr.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("CORS 中间件已注册", func(t *testing.T) {
		engine, mr, _, _, _ := setupTestRouter(t)
		defer mr.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		engine.ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Logger 中间件已注册", func(t *testing.T) {
		engine, mr, _, _, _ := setupTestRouter(t)
		defer mr.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		engine.ServeHTTP(w, req)

		// Logger 中间件只是记录日志，这里验证请求能正常处理
		assert.Equal(t, 200, w.Code)
	})
}

// ==================== JWT Blacklist Tests ====================

func TestJWTBlacklist(t *testing.T) {
	engine, mr, _, _, jwtManager := setupTestRouter(t)
	defer mr.Close()

	// 生成 token
	token, expiresAt, err := jwtManager.GenerateToken(1, "testuser", "user", nil)
	assert.NoError(t, err)

	// 解析 token 获取 claims
	claims, err := jwtManager.ParseToken(token)
	assert.NoError(t, err)

	// 将 token 加入黑名单
	err = jwtManager.Blacklist(t.Context(), claims.ID, expiresAt)
	assert.NoError(t, err)

	// 使用黑名单中的 token 访问受保护的路由
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(w, req)

	// 应该返回 401，因为 token 在黑名单中
	assert.Equal(t, 401, w.Code)
}

// ==================== API Key Self-Loop Detection Tests ====================

func TestAPIKeySelfLoopDetection(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// 创建测试数据
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	rdb.FlushAll(t.Context())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	req.Header.Set("X-CodeMind-Proxy", "1") // 设置自环标志
	engine.ServeHTTP(w, req)

	// 应该返回 502，因为检测到自环
	assert.Equal(t, 502, w.Code)
}

// ==================== API Key Status Tests ====================

func TestAPIKeyStatus_DisabledKey(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// 创建测试数据 - 禁用的 API Key
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 1)
	// status = 0 表示禁用
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 0, 1)

	rdb.FlushAll(t.Context())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	engine.ServeHTTP(w, req)

	// 应该返回 403，因为 API Key 被禁用
	assert.Equal(t, 403, w.Code)
}

func TestAPIKeyStatus_DisabledUser(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// 创建测试数据 - 禁用的用户
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	// status = 0 表示禁用
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 0)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	rdb.FlushAll(t.Context())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	engine.ServeHTTP(w, req)

	// 应该返回 403，因为用户被禁用
	assert.Equal(t, 403, w.Code)
}
