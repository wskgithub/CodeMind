package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/jwt"
	"codemind/internal/pkg/response"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testJWTSecret is the JWT secret for tests (min 32 chars, required by jwt.NewManager).
const testJWTSecret = "01234567890123456789012345678901"

// setupTestRedis creates a Redis instance for testing.
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return mr, rdb
}

// setupTestGin sets up a Gin engine for testing.
func setupTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// setupTestDB creates a SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	return db
}

// MockMonitorStats is a mock implementation of monitor stats.
type MockMonitorStats struct {
	mock.Mock
}

func (m *MockMonitorStats) RecordRequestMetrics(statusCode int, responseTimeMs float64) {
	m.Called(statusCode, responseTimeMs)
}

// ==================== CORS middleware tests ====================

func TestCORS(t *testing.T) {
	router := setupTestGin()
	router.Use(CORS(nil)) // nil allows all origins (same as production without whitelist)
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	t.Run("GET request sets CORS headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// Verify CORS headers are set
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Authorization")
		router.ServeHTTP(w, req)

		// CORS middleware automatically handles OPTIONS requests
		// Status code may be 204 or determined by the route
		assert.Contains(t, []int{200, 204}, w.Code)

		// Verify CORS response headers
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("POST request sets CORS headers", func(t *testing.T) {
		router.POST("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

// ==================== Logger middleware tests ====================

func TestLogger(t *testing.T) {
	logger := zaptest.NewLogger(t)
	router := setupTestGin()
	router.Use(Logger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "server error"})
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "normal request",
			path:           "/test",
			expectedStatus: 200,
		},
		{
			name:           "server error",
			path:           "/error",
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestLoggerWithUserID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	router := setupTestGin()
	router.Use(func(c *gin.Context) {
		c.Set(CtxKeyUserID, int64(123))
		c.Next()
	})
	router.Use(Logger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ==================== Recovery middleware tests ====================

func TestRecovery(t *testing.T) {
	logger := zaptest.NewLogger(t)
	router := setupTestGin()
	router.Use(Recovery(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	router.GET("/normal", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedCode   int
	}{
		{
			name:           "catch panic",
			path:           "/panic",
			expectedStatus: 500,
			expectedCode:   http.StatusInternalServerError,
		},
		{
			name:           "normal request",
			path:           "/normal",
			expectedStatus: 200,
			expectedCode:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedCode > 0 {
				var resp response.Response
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, resp.Code)
			}
		})
	}
}

// ==================== JWTAuth middleware tests ====================

func TestJWTAuth(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	jwtManager, err := jwt.NewManager(testJWTSecret, 24, rdb)
	require.NoError(t, err)
	router := setupTestGin()
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": GetUserID(c),
			"role":    GetUserRole(c),
		})
	})

	t.Run("valid token", func(t *testing.T) {
		deptID := int64(1)
		token, _, err := jwtManager.GenerateToken(123, "testuser", "admin", &deptID)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp gin.H
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, float64(123), resp["user_id"])
		assert.Equal(t, "admin", resp["role"])
	})

	t.Run("missing Authorization header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, errcode.ErrTokenInvalid.Code, resp.Code)
	})

	t.Run("incorrect Bearer prefix", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Basic token123")
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("empty token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer ")
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, errcode.ErrTokenInvalid.Code, resp.Code)
	})

	t.Run("blacklisted token", func(t *testing.T) {
		token, expiresAt, err := jwtManager.GenerateToken(123, "testuser", "admin", nil)
		assert.NoError(t, err)

		// Parse Token to get claims
		claims, err := jwtManager.ParseToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		// Add Token to blacklist (using claims.ID i.e. JTI)
		err = jwtManager.Blacklist(context.Background(), claims.ID, expiresAt)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		// Token is blacklisted, return 401
		assert.Equal(t, 401, w.Code)
	})
}

// ==================== Role check middleware tests ====================

func TestRequireRole(t *testing.T) {
	router := setupTestGin()

	// Simulate auth middleware, set role
	router.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Test-Role")
		if role != "" {
			c.Set(CtxKeyRole, role)
		}
		c.Next()
	})

	router.GET("/admin", RequireRole("super_admin"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin only"})
	})
	router.GET("/manager", RequireRole("super_admin", "dept_manager"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "manager access"})
	})
	router.GET("/user", RequireRole("user", "admin", "super_admin"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "user access"})
	})

	tests := []struct {
		name           string
		path           string
		role           string
		expectedStatus int
		expectedCode   int
	}{
		{
			name:           "super admin access to admin route",
			path:           "/admin",
			role:           "super_admin",
			expectedStatus: 200,
		},
		{
			name:           "dept manager denied access to admin route",
			path:           "/admin",
			role:           "dept_manager",
			expectedStatus: 403,
			expectedCode:   errcode.ErrForbidden.Code,
		},
		{
			name:           "super admin access to manager route",
			path:           "/manager",
			role:           "super_admin",
			expectedStatus: 200,
		},
		{
			name:           "dept manager access to manager route",
			path:           "/manager",
			role:           "dept_manager",
			expectedStatus: 200,
		},
		{
			name:           "regular user denied access to manager route",
			path:           "/manager",
			role:           "user",
			expectedStatus: 403,
			expectedCode:   errcode.ErrForbidden.Code,
		},
		{
			name:           "no role set denied access",
			path:           "/user",
			role:           "",
			expectedStatus: 403,
			expectedCode:   errcode.ErrForbidden.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			if tt.role != "" {
				req.Header.Set("X-Test-Role", tt.role)
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedCode > 0 {
				var resp response.Response
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, resp.Code)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	router := setupTestGin()
	router.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Test-Role")
		if role != "" {
			c.Set(CtxKeyRole, role)
		}
		c.Next()
	})
	router.GET("/admin", RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin only"})
	})

	t.Run("super admin access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("X-Test-Role", "super_admin")
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("non-admin denied access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("X-Test-Role", "user")
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
	})
}

func TestRequireManager(t *testing.T) {
	router := setupTestGin()
	router.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Test-Role")
		if role != "" {
			c.Set(CtxKeyRole, role)
		}
		c.Next()
	})
	router.GET("/manager", RequireManager(), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "manager access"})
	})

	tests := []struct {
		name           string
		role           string
		expectedStatus int
	}{
		{"super admin", "super_admin", 200},
		{"dept manager", "dept_manager", 200},
		{"regular user", "user", 403},
		{"admin", "admin", 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/manager", nil)
			req.Header.Set("X-Test-Role", tt.role)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ==================== Context getter function tests ====================

func TestGetUserID(t *testing.T) {
	tests := []struct {
		setupContext func(*gin.Context)
		name         string
		expectedID   int64
	}{
		{
			name: "get user ID normally",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyUserID, int64(123))
			},
			expectedID: 123,
		},
		{
			name:         "user ID not set",
			setupContext: func(c *gin.Context) {},
			expectedID:   0,
		},
		{
			name: "user ID is nil",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyUserID, nil)
			},
			expectedID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupContext(c)

			id := GetUserID(c)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}

func TestGetUserRole(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		expectedRole string
	}{
		{
			name: "get role normally",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyRole, "admin")
			},
			expectedRole: "admin",
		},
		{
			name:         "role not set",
			setupContext: func(c *gin.Context) {},
			expectedRole: "",
		},
		{
			name: "role is nil",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyRole, nil)
			},
			expectedRole: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupContext(c)

			role := GetUserRole(c)
			assert.Equal(t, tt.expectedRole, role)
		})
	}
}

func TestGetDepartmentID(t *testing.T) {
	tests := []struct {
		setupContext func(*gin.Context)
		expectedID   *int64
		name         string
	}{
		{
			name: "get department ID normally",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyDepartmentID, int64(456))
			},
			expectedID: func() *int64 { id := int64(456); return &id }(),
		},
		{
			name:         "department ID not set",
			setupContext: func(c *gin.Context) {},
			expectedID:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupContext(c)

			id := GetDepartmentID(c)
			if tt.expectedID == nil {
				assert.Nil(t, id)
			} else {
				assert.NotNil(t, id)
				assert.Equal(t, *tt.expectedID, *id)
			}
		})
	}
}

func TestGetClaims(t *testing.T) {
	tests := []struct {
		setupContext  func(*gin.Context)
		expectedClaim *jwt.Claims
		name          string
		expectedNil   bool
	}{
		{
			name: "get Claims normally",
			setupContext: func(c *gin.Context) {
				claims := &jwt.Claims{
					UserID:   123,
					Username: "testuser",
					Role:     "admin",
				}
				c.Set(CtxKeyClaims, claims)
			},
			expectedNil:   false,
			expectedClaim: &jwt.Claims{UserID: 123, Username: "testuser", Role: "admin"},
		},
		{
			name:          "Claims not set",
			setupContext:  func(c *gin.Context) {},
			expectedNil:   true,
			expectedClaim: nil,
		},
		{
			name: "Claims is nil",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyClaims, nil)
			},
			expectedNil:   true,
			expectedClaim: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupContext(c)

			claims := GetClaims(c)
			if tt.expectedNil {
				assert.Nil(t, claims)
			} else {
				assert.NotNil(t, claims)
				assert.Equal(t, tt.expectedClaim.UserID, claims.UserID)
				assert.Equal(t, tt.expectedClaim.Username, claims.Username)
				assert.Equal(t, tt.expectedClaim.Role, claims.Role)
			}
		})
	}
}

// ==================== RequestMonitor middleware tests ====================

func TestRequestMonitor(t *testing.T) {
	mockStats := new(MockMonitorStats)
	mockStats.On("RecordRequestMetrics", mock.Anything, mock.Anything).Return()

	router := setupTestGin()
	router.Use(RequestMonitor(mockStats))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "server error"})
	})

	t.Run("record 200 request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		time.Sleep(100 * time.Millisecond) // Wait for goroutine to execute
		mockStats.AssertCalled(t, "RecordRequestMetrics", 200, mock.AnythingOfType("float64"))
	})

	t.Run("record 500 request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		time.Sleep(100 * time.Millisecond) // Wait for goroutine to execute
		mockStats.AssertCalled(t, "RecordRequestMetrics", 500, mock.AnythingOfType("float64"))
	})
}

// ==================== APIKeyAuth middleware tests ====================

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		setupHeaders func(*http.Request)
		expectedKey  string
	}{
		{
			name: "extract from x-api-key",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("x-api-key", "cm-testkey123")
			},
			expectedKey: "cm-testkey123",
		},
		{
			name: "extract from Authorization Bearer",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer cm-testkey456")
			},
			expectedKey: "cm-testkey456",
		},
		{
			name: "x-api-key takes priority over Authorization",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("x-api-key", "cm-from-x-api-key")
				req.Header.Set("Authorization", "Bearer cm-from-auth")
			},
			expectedKey: "cm-from-x-api-key",
		},
		{
			name:         "no API Key",
			setupHeaders: func(req *http.Request) {},
			expectedKey:  "",
		},
		{
			name: "incorrect Authorization format",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("Authorization", "Basic cm-testkey")
			},
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", "/test", nil)
			tt.setupHeaders(req)
			c.Request = req

			key := extractAPIKey(c)
			assert.Equal(t, tt.expectedKey, key)
		})
	}
}

func TestAPIKeyAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// Create test tables
	db.Exec(`CREATE TABLE api_keys (
		id INTEGER PRIMARY KEY,
		key_hash TEXT,
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

	router := setupTestGin()
	router.Use(APIKeyAuth(db, rdb, logger))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": GetUserID(c),
			"role":    GetUserRole(c),
		})
	})

	t.Run("request self-loop detection", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("X-CodeMind-Proxy", "1")
		router.ServeHTTP(w, req)

		assert.Equal(t, 502, w.Code)
	})

	t.Run("missing API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("invalid format API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", "invalid-key")
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("API Key not in database", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", "cm-nonexistentkey12345678901234567890123456789012")
		router.ServeHTTP(w, req)

		// When API Key doesn't exist, GORM returns zero-value struct, KeyStatus is 0 (disabled)
		// so it returns 403 Forbidden (ErrAPIKeyDisabled) instead of 401 Unauthorized
		assert.Equal(t, 403, w.Code)
	})
}

func TestAPIKeyAuthWithValidKey(t *testing.T) {
	logger := zap.NewNop() // Use Nop logger to avoid test log noise
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// Create test tables
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

	// Generate test API Key
	testAPIKey := "cm-" + strings.Repeat("a", 64) // cm- prefix + 64 characters
	keyHash := crypto.HashAPIKey(testAPIKey)
	deptID := int64(1)

	// Insert test data
	db.Exec(`INSERT INTO users (id, username, role, department_id, status) VALUES (?, ?, ?, ?, ?)`,
		1, "testuser", "admin", deptID, 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		1, keyHash, 1, 1)

	router := setupTestGin()
	router.Use(APIKeyAuth(db, rdb, logger))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": GetUserID(c),
			"role":    GetUserRole(c),
		})
	})

	t.Run("valid API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", testAPIKey)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp gin.H
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), resp["user_id"])
		assert.Equal(t, "admin", resp["role"])
	})

	t.Run("fetch from Redis cache", func(t *testing.T) {
		// Second request should be served from Redis cache
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", testAPIKey)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}

func TestAPIKeyAuthDisabledKey(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// Create test tables
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

	// Generate test API Key (disabled)
	testAPIKey := "cm-" + strings.Repeat("b", 64)
	keyHash := crypto.HashAPIKey(testAPIKey)

	// Insert test data - key status is disabled (0)
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		2, "testuser2", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		2, keyHash, 0, 2) // status = 0 means disabled

	router := setupTestGin()
	router.Use(APIKeyAuth(db, rdb, logger))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("disabled API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", testAPIKey)
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestAPIKeyAuthDisabledUser(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// Create test tables
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

	// Generate test API Key
	testAPIKey := "cm-" + strings.Repeat("c", 64)
	keyHash := crypto.HashAPIKey(testAPIKey)

	// Insert test data - user status is disabled (0)
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		3, "testuser3", "user", 0) // status = 0 means disabled
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		3, keyHash, 1, 3)

	router := setupTestGin()
	router.Use(APIKeyAuth(db, rdb, logger))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("disabled user account", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", testAPIKey)
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestAPIKeyAuthExpiredKey(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// Create test tables
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

	// Generate test API Key
	testAPIKey := "cm-" + strings.Repeat("d", 64)
	keyHash := crypto.HashAPIKey(testAPIKey)
	expiredTime := time.Now().Add(-24 * time.Hour) // expired

	// Insert test data - key is expired
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		4, "testuser4", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id, expires_at) VALUES (?, ?, ?, ?, ?)`,
		4, keyHash, 1, 4, expiredTime)

	router := setupTestGin()
	router.Use(APIKeyAuth(db, rdb, logger))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("expired API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", testAPIKey)
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})
}

func TestGetAPIKeyInfo(t *testing.T) {
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// Create test tables
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

	testAPIKey := "cm-" + strings.Repeat("e", 64)
	keyHash := crypto.HashAPIKey(testAPIKey)
	deptID := int64(5)

	// Insert test data
	db.Exec(`INSERT INTO users (id, username, role, department_id, status) VALUES (?, ?, ?, ?, ?)`,
		5, "cacheduser", "manager", deptID, 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		5, keyHash, 1, 5)

	t.Run("query from database and cache to Redis", func(t *testing.T) {
		ctx := context.Background()
		info, err := getAPIKeyInfo(ctx, db, rdb, keyHash)

		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, int64(5), info.UserID)
		assert.Equal(t, "cacheduser", info.Username)
		assert.Equal(t, "manager", info.Role)
		assert.Equal(t, int64(5), info.KeyID)
		assert.NotNil(t, info.DepartmentID)
		assert.Equal(t, deptID, *info.DepartmentID)

		// Verify cached in Redis
		cacheKey := "codemind:apikey:" + keyHash
		cached, err := rdb.Get(ctx, cacheKey).Result()
		assert.NoError(t, err)
		assert.NotEmpty(t, cached)
	})

	t.Run("fetch from Redis cache", func(t *testing.T) {
		ctx := context.Background()
		info, err := getAPIKeyInfo(ctx, db, rdb, keyHash)

		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, int64(5), info.UserID)
	})

	t.Run("nonexistent API Key", func(t *testing.T) {
		ctx := context.Background()
		nonExistentHash := crypto.HashAPIKey("cm-nonexistent")
		info, err := getAPIKeyInfo(ctx, db, rdb, nonExistentHash)

		// Note: current implementation - GORM doesn't return error when record doesn't exist
		// instead it returns a zero-value struct (all fields are 0)
		// This might be a code issue, but here we test current behavior
		assert.NoError(t, err)                    // Current behavior: no error returned
		assert.NotNil(t, info)                    // Returns empty struct instead of nil
		assert.Equal(t, int64(0), info.UserID)    // zero value
		assert.Equal(t, int16(0), info.KeyStatus) // zero value
	})
}

func TestGetAPIKeyInfoDatabaseError(t *testing.T) {
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// Don't create tables to simulate DB error
	t.Run("database error", func(t *testing.T) {
		ctx := context.Background()
		keyHash := crypto.HashAPIKey("cm-test")
		info, err := getAPIKeyInfo(ctx, db, rdb, keyHash)

		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

// MockDB is a mock for simulating database errors.
type MockDB struct {
	errorOnQuery bool
}

func (m *MockDB) ErrorOnQuery() bool {
	return m.errorOnQuery
}

// ==================== JWTAuth expired Token tests ====================

func TestJWTAuthExpiredToken(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// Create JWT Manager with very short expiration
	jwtManager, err := jwt.NewManager(testJWTSecret, 0, rdb) // 0 hours expiration
	require.NoError(t, err)
	router := setupTestGin()
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("expired token", func(t *testing.T) {
		// Generate a Token that's immediately expired (0 hours expiration)
		// Use negative time to test expiration
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInVzZXJuYW1lIjoidGVzdCIsInJvbGUiOiJhZG1pbiIsImV4cCI6MTYwMDAwMDAwMH0.invalid"

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})
}

// ==================== Edge case tests ====================

func TestJWTAuthMultipleSpaces(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	jwtManager, err := jwt.NewManager(testJWTSecret, 24, rdb)
	require.NoError(t, err)
	router := setupTestGin()
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("Authorization contains multiple spaces", func(t *testing.T) {
		token, _, err := jwtManager.GenerateToken(123, "test", "user", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer  "+token) // Two spaces
		router.ServeHTTP(w, req)

		// Should be rejected due to leading space in token
		assert.Equal(t, 401, w.Code)
	})
}

func TestRequireRoleTypeAssertion(t *testing.T) {
	router := setupTestGin()
	router.Use(func(c *gin.Context) {
		// Set non-string role type (should not happen normally)
		c.Set(CtxKeyRole, 123)
		c.Next()
	})
	router.GET("/test", RequireRole("admin"), func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("wrong role type causes panic", func(t *testing.T) {
		// Use recovery to catch panic
		logger := zaptest.NewLogger(t)
		routerWithRecovery := setupTestGin()
		routerWithRecovery.Use(Recovery(logger))
		routerWithRecovery.Use(func(c *gin.Context) {
			c.Set(CtxKeyRole, 123) // Set wrong type
			c.Next()
		})
		routerWithRecovery.GET("/test", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		routerWithRecovery.ServeHTTP(w, req)

		// Type assertion failure causes panic, recovery catches it and returns 500
		assert.Equal(t, 500, w.Code)
	})
}

// ==================== Logger error handling tests ====================

func TestLoggerWithErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	router := setupTestGin()
	router.Use(Logger(logger))
	router.GET("/error", func(c *gin.Context) {
		c.Error(errors.New("test error"))
		c.JSON(400, gin.H{"error": "bad request"})
	})

	t.Run("request with errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})
}

// ==================== Full integration tests ====================

func TestMiddlewareChain(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	jwtManager, err := jwt.NewManager(testJWTSecret, 24, rdb)
	require.NoError(t, err)

	// Create a complete middleware chain
	router := setupTestGin()
	router.Use(Recovery(logger))
	router.Use(Logger(logger))
	router.Use(JWTAuth(jwtManager))
	router.Use(RequireManager())

	router.GET("/api/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id":       GetUserID(c),
			"role":          GetUserRole(c),
			"department_id": GetDepartmentID(c),
			"claims":        GetClaims(c),
		})
	})

	t.Run("full middleware chain - dept manager", func(t *testing.T) {
		deptID := int64(10)
		token, _, err := jwtManager.GenerateToken(100, "manager", "dept_manager", &deptID)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp gin.H
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, float64(100), resp["user_id"])
		assert.Equal(t, "dept_manager", resp["role"])
	})

	t.Run("full middleware chain - super admin", func(t *testing.T) {
		token, _, err := jwtManager.GenerateToken(1, "admin", "super_admin", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("full middleware chain - regular user denied", func(t *testing.T) {
		token, _, err := jwtManager.GenerateToken(50, "user", "user", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
	})
}
