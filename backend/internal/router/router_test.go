package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"codemind/internal/model"
	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/jwt"

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

// testJWTSecret is the JWT secret for router tests (min 32 chars, required by jwt.NewManager).
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

	// Create required tables for API Key auth tests
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
	// Test route registration and middleware with empty handlers
	// Note: empty handlers cause 500, but useful for verifying route existence and middleware
	handlers := &Handlers{} // empty handlers
	// nil corsOrigins means AllowAllOrigins, Access-Control-Allow-Origin is *
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
		expectExists bool // whether the route should exist (even without handler implementation)
	}{
		{
			name:         "health check",
			method:       "GET",
			path:         "/health",
			expectExists: true,
		},
		{
			name:         "login endpoint",
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

			// Existing route should not return 404
			if tt.expectExists {
				assert.NotEqual(t, 404, w.Code, "route should exist")
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
			name:           "unauthenticated access to logout",
			method:         "POST",
			path:           "/api/v1/auth/logout",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to profile",
			method:         "GET",
			path:           "/api/v1/auth/profile",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to update profile",
			method:         "PUT",
			path:           "/api/v1/auth/profile",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to change password",
			method:         "PUT",
			path:           "/api/v1/auth/password",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to keys",
			method:         "GET",
			path:           "/api/v1/keys",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to users",
			method:         "GET",
			path:           "/api/v1/users",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to departments",
			method:         "GET",
			path:           "/api/v1/departments",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to stats overview",
			method:         "GET",
			path:           "/api/v1/stats/overview",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to limits my",
			method:         "GET",
			path:           "/api/v1/limits/my",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to announcements",
			method:         "GET",
			path:           "/api/v1/announcements",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to system configs",
			method:         "GET",
			path:           "/api/v1/system/configs",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to mcp services",
			method:         "GET",
			path:           "/api/v1/mcp/services",
			expectedStatus: 401,
			expectedCode:   errcode.ErrTokenInvalid.Code,
		},
		{
			name:           "unauthenticated access to monitor dashboard",
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

	// Generate valid token
	token, _, err := jwtManager.GenerateToken(1, "testuser", "user", nil)
	assert.NoError(t, err)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "authenticated access to logout",
			method: "POST",
			path:   "/api/v1/auth/logout",
		},
		{
			name:   "authenticated access to profile",
			method: "GET",
			path:   "/api/v1/auth/profile",
		},
		{
			name:   "authenticated access to update profile",
			method: "PUT",
			path:   "/api/v1/auth/profile",
		},
		{
			name:   "authenticated access to change password",
			method: "PUT",
			path:   "/api/v1/auth/password",
		},
		{
			name:   "authenticated access to keys",
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

			// Route exists and auth passed, should not return 401
			assert.NotEqual(t, 401, w.Code, "should not return 401 since token is valid")
			assert.NotEqual(t, 404, w.Code, "route should exist")
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
			name:           "invalid token format",
			token:          "invalid-token",
			expectedStatus: 401,
		},
		{
			name:           "incorrect Bearer format",
			token:          "Basic invalid",
			expectedStatus: 401,
		},
		{
			name:           "empty Bearer token",
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
			name:           "super admin access to system configs",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/system/configs",
			method:         "GET",
			expectedStatus: 500, // Route matched, handler is nil, Recovery middleware returns 500
		},
		{
			name:           "dept manager denied access to system configs",
			role:           model.RoleDeptManager,
			path:           "/api/v1/system/configs",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "regular user denied access to system configs",
			role:           model.RoleUser,
			path:           "/api/v1/system/configs",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "super admin access to MCP management",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/mcp/services",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "dept manager denied access to MCP management",
			role:           model.RoleDeptManager,
			path:           "/api/v1/mcp/services",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "super admin access to monitoring",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/monitor/dashboard",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "dept manager denied access to monitoring",
			role:           model.RoleDeptManager,
			path:           "/api/v1/monitor/dashboard",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "super admin access to LLM backend management",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/system/llm-backends",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "dept manager denied access to LLM backend management",
			role:           model.RoleDeptManager,
			path:           "/api/v1/system/llm-backends",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "super admin access to audit logs",
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
			name:           "super admin access to user management",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/users",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "dept manager access to user management",
			role:           model.RoleDeptManager,
			path:           "/api/v1/users",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "regular user denied access to user management",
			role:           model.RoleUser,
			path:           "/api/v1/users",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "super admin access to department management",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/departments",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "dept manager access to department management",
			role:           model.RoleDeptManager,
			path:           "/api/v1/departments",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "regular user denied access to department management",
			role:           model.RoleUser,
			path:           "/api/v1/departments",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "super admin access to rate limit management",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/limits",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "dept manager access to rate limit management",
			role:           model.RoleDeptManager,
			path:           "/api/v1/limits",
			method:         "GET",
			expectedStatus: 500,
		},
		{
			name:           "regular user denied access to rate limit management",
			role:           model.RoleUser,
			path:           "/api/v1/limits",
			method:         "GET",
			expectedStatus: 403,
		},
		{
			name:           "super admin export CSV",
			role:           model.RoleSuperAdmin,
			path:           "/api/v1/stats/export/csv",
			method:         "GET",
			expectedStatus: 400, // handler is nil but param validation returns 400
		},
		{
			name:           "dept manager export CSV",
			role:           model.RoleDeptManager,
			path:           "/api/v1/stats/export/csv",
			method:         "GET",
			expectedStatus: 400, // handler is nil but param validation returns 400
		},
		{
			name:           "regular user denied export CSV",
			role:           model.RoleUser,
			path:           "/api/v1/stats/export/csv",
			method:         "GET",
			expectedStatus: 400, // Params checked first, then permissions
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
			name:           "no API Key access to chat completions",
			method:         "POST",
			path:           "/api/openai/v1/chat/completions",
			expectedStatus: 401,
		},
		{
			name:           "no API Key access to completions",
			method:         "POST",
			path:           "/api/openai/v1/completions",
			expectedStatus: 401,
		},
		{
			name:           "no API Key access to models",
			method:         "GET",
			path:           "/api/openai/v1/models",
			expectedStatus: 401,
		},
		{
			name:           "no API Key access to embeddings",
			method:         "POST",
			path:           "/api/openai/v1/embeddings",
			expectedStatus: 401,
		},
		{
			name:           "no API Key access to messages",
			method:         "POST",
			path:           "/api/anthropic/v1/messages",
			expectedStatus: 401,
		},
		{
			name:           "no API Key access to responses",
			method:         "POST",
			path:           "/api/openai/v1/responses",
			expectedStatus: 401,
		},
		{
			name:           "no API Key access to MCP SSE",
			method:         "GET",
			path:           "/mcp/sse",
			expectedStatus: 401,
		},
		{
			name:           "no API Key access to MCP message",
			method:         "POST",
			path:           "/mcp/message",
			expectedStatus: 401,
		},
		{
			name:           "no API Key access to MCP Streamable HTTP",
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
			name:           "invalid API Key format",
			apiKey:         "invalid-key",
			expectedStatus: 401,
		},
		{
			name:           "API Key without cm- prefix",
			apiKey:         "sk-testkey123",
			expectedStatus: 401,
		},
		{
			name:           "empty API Key",
			apiKey:         "",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/openai/v1/chat/completions", nil)
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

	// Create test data
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	// Insert test user and API Key
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	// Clear Redis cache
	rdb.FlushAll(t.Context())

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int // 500 means route and middleware passed but handler not implemented
	}{
		{
			name:           "valid API Key access to chat completions",
			method:         "POST",
			path:           "/api/openai/v1/chat/completions",
			expectedStatus: 500,
		},
		{
			name:           "valid API Key access to completions",
			method:         "POST",
			path:           "/api/openai/v1/completions",
			expectedStatus: 500,
		},
		{
			name:           "valid API Key access to models",
			method:         "GET",
			path:           "/api/openai/v1/models",
			expectedStatus: 500,
		},
		{
			name:           "valid API Key access to embeddings",
			method:         "POST",
			path:           "/api/openai/v1/embeddings",
			expectedStatus: 500,
		},
		{
			name:           "valid API Key access to messages",
			method:         "POST",
			path:           "/api/anthropic/v1/messages",
			expectedStatus: 500,
		},
		{
			name:           "valid API Key access to MCP SSE",
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

			// Should not return 401 (auth passed) or 404 (route exists)
			assert.NotEqual(t, 401, w.Code, "should not return 401 since API Key is valid")
			assert.NotEqual(t, 404, w.Code, "route should exist")
		})
	}
}

func TestAPIKeyProtectedRoutes_WithXAPIKey(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// Create test data
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	rdb.FlushAll(t.Context())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/openai/v1/models", nil)
	// Use x-api-key header (Anthropic format)
	req.Header.Set("x-api-key", testAPIKey)
	engine.ServeHTTP(w, req)

	// Should not return 401 since middleware supports x-api-key
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

	// CORS middleware should handle OPTIONS requests
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
			name:   "access nonexistent path",
			path:   "/nonexistent",
			method: "GET",
		},
		{
			name:   "access nonexistent API path",
			path:   "/api/v999/test",
			method: "GET",
		},
		{
			name:   "access nonexistent auth path",
			path:   "/api/v1/auth/nonexistent",
			method: "GET",
		},
		{
			name:   "access nonexistent v1 path",
			path:   "/v1/nonexistent",
			method: "GET",
		},
		{
			name:   "access nonexistent mcp path",
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

	// Create valid token
	token, _, err := jwtManager.GenerateToken(1, "testuser", model.RoleSuperAdmin, nil)
	assert.NoError(t, err)

	// Test all major routes are properly registered
	routes := []struct {
		path          string
		method        string
		requireAuth   bool
		requireAPIKey bool
	}{
		// Public routes
		{"/health", "GET", false, false},
		{"/api/v1/auth/login", "POST", false, false},

		// JWT protected routes - auth
		{"/api/v1/auth/logout", "POST", true, false},
		{"/api/v1/auth/profile", "GET", true, false},
		{"/api/v1/auth/profile", "PUT", true, false},
		{"/api/v1/auth/password", "PUT", true, false},

		// JWT protected routes - API Key
		{"/api/v1/keys", "GET", true, false},
		{"/api/v1/keys", "POST", true, false},

		// JWT protected routes - user management
		{"/api/v1/users", "GET", true, false},
		{"/api/v1/users/1", "GET", true, false},

		// JWT protected routes - department management
		{"/api/v1/departments", "GET", true, false},
		{"/api/v1/departments/1", "GET", true, false},

		// JWT protected routes - stats
		{"/api/v1/stats/overview", "GET", true, false},
		{"/api/v1/stats/usage", "GET", true, false},
		{"/api/v1/stats/ranking", "GET", true, false},

		// JWT protected routes - rate limits
		{"/api/v1/limits/my", "GET", true, false},
		{"/api/v1/limits/my/progress", "GET", true, false},

		// JWT protected routes - announcements
		{"/api/v1/announcements", "GET", true, false},

		// JWT protected routes - system management (admin required)
		{"/api/v1/system/configs", "GET", true, false},
		{"/api/v1/system/llm-backends", "GET", true, false},
		{"/api/v1/system/audit-logs", "GET", true, false},

		// JWT protected routes - MCP management (admin required)
		{"/api/v1/mcp/services", "GET", true, false},
		{"/api/v1/mcp/access-rules", "GET", true, false},

		// JWT protected routes - monitoring (admin required)
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

			// Route exists (should not return 404)
			assert.NotEqual(t, 404, w.Code, "route should exist")
		})
	}
}

// ==================== Middleware Chain Tests ====================

func TestMiddlewareChain(t *testing.T) {
	t.Run("Recovery middleware registered", func(t *testing.T) {
		engine, mr, _, _, _ := setupTestRouter(t)
		defer mr.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("CORS middleware registered", func(t *testing.T) {
		engine, mr, _, _, _ := setupTestRouter(t)
		defer mr.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		engine.ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Logger middleware registered", func(t *testing.T) {
		engine, mr, _, _, _ := setupTestRouter(t)
		defer mr.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		engine.ServeHTTP(w, req)

		// Logger middleware only logs; verify request is handled normally
		assert.Equal(t, 200, w.Code)
	})
}

// ==================== JWT Blacklist Tests ====================

func TestJWTBlacklist(t *testing.T) {
	engine, mr, _, _, jwtManager := setupTestRouter(t)
	defer mr.Close()

	// Generate token
	token, expiresAt, err := jwtManager.GenerateToken(1, "testuser", "user", nil)
	assert.NoError(t, err)

	// Parse token to get claims
	claims, err := jwtManager.ParseToken(token)
	assert.NoError(t, err)

	// Add token to blacklist
	err = jwtManager.Blacklist(t.Context(), claims.ID, expiresAt)
	assert.NoError(t, err)

	// Access protected route with blacklisted token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(w, req)

	// Should return 401 since token is blacklisted
	assert.Equal(t, 401, w.Code)
}

// ==================== API Key Self-Loop Detection Tests ====================

func TestAPIKeySelfLoopDetection(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// Create test data
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	rdb.FlushAll(t.Context())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/openai/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	req.Header.Set("X-CodeMind-Proxy", "1") // Set self-loop flag
	engine.ServeHTTP(w, req)

	// Should return 502 due to self-loop detection
	assert.Equal(t, 502, w.Code)
}

// ==================== API Key Status Tests ====================

func TestAPIKeyStatus_DisabledKey(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// Create test data - disabled API Key
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 1)
	// status = 0 means disabled
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 0, 1)

	rdb.FlushAll(t.Context())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/openai/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	engine.ServeHTTP(w, req)

	// Should return 403 since API Key is disabled
	assert.Equal(t, 403, w.Code)
}

func TestAPIKeyStatus_DisabledUser(t *testing.T) {
	engine, mr, rdb, db, _ := setupTestRouter(t)
	defer mr.Close()

	// Create test data - disabled user
	testAPIKey := "cm-testapikey123456789012345678901234567890123456789012345678901234"
	keyHash := crypto.HashAPIKey(testAPIKey)

	// status = 0 means disabled
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		1, "testuser", "user", 0)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	rdb.FlushAll(t.Context())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/openai/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	engine.ServeHTTP(w, req)

	// Should return 403 since user is disabled
	assert.Equal(t, 403, w.Code)
}
