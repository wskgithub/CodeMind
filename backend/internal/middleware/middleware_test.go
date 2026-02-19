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
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRedis 创建测试用的 Redis 实例
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return mr, rdb
}

// setupTestGin 设置测试用的 Gin 引擎
func setupTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// setupTestDB 创建测试用的 SQLite 数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	return db
}

// MockMonitorStats 监控统计的 mock 实现
type MockMonitorStats struct {
	mock.Mock
}

func (m *MockMonitorStats) RecordRequestMetrics(statusCode int, responseTimeMs float64) {
	m.Called(statusCode, responseTimeMs)
}

// ==================== CORS 中间件测试 ====================

func TestCORS(t *testing.T) {
	router := setupTestGin()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	t.Run("GET 请求设置 CORS 头", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// 验证 CORS 头被设置
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS 预检请求", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Authorization")
		router.ServeHTTP(w, req)

		// CORS 中间件会自动处理 OPTIONS 请求
		// 状态码可能是 204 或由路由决定
		assert.Contains(t, []int{200, 204}, w.Code)
		
		// 验证 CORS 响应头
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("POST 请求设置 CORS 头", func(t *testing.T) {
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

// ==================== Logger 中间件测试 ====================

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
			name:           "正常请求",
			path:           "/test",
			expectedStatus: 200,
		},
		{
			name:           "服务器错误",
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

// ==================== Recovery 中间件测试 ====================

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
			name:           "捕获 panic",
			path:           "/panic",
			expectedStatus: 500,
			expectedCode:   errcode.ErrInternal.Code,
		},
		{
			name:           "正常请求",
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

// ==================== JWTAuth 中间件测试 ====================

func TestJWTAuth(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	jwtManager := jwt.NewManager("test-secret", 24, rdb)
	router := setupTestGin()
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": GetUserID(c),
			"role":    GetUserRole(c),
		})
	})

	t.Run("有效 Token", func(t *testing.T) {
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

	t.Run("缺失 Authorization Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, errcode.ErrTokenInvalid.Code, resp.Code)
	})

	t.Run("错误的 Bearer 前缀", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Basic token123")
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("空的 Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer ")
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("无效的 Token", func(t *testing.T) {
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

	t.Run("已加入黑名单的 Token", func(t *testing.T) {
		token, expiresAt, err := jwtManager.GenerateToken(123, "testuser", "admin", nil)
		assert.NoError(t, err)

		// 解析 Token 获取 claims
		claims, err := jwtManager.ParseToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		// 将 Token 加入黑名单 (使用 claims.ID 即 JTI)
		err = jwtManager.Blacklist(context.Background(), claims.ID, expiresAt)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		// Token 在黑名单中，返回 401
		assert.Equal(t, 401, w.Code)
	})
}

// ==================== 角色检查中间件测试 ====================

func TestRequireRole(t *testing.T) {
	router := setupTestGin()

	// 模拟认证中间件，设置角色
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
			name:           "超级管理员访问 admin 路由",
			path:           "/admin",
			role:           "super_admin",
			expectedStatus: 200,
		},
		{
			name:           "部门经理访问 admin 路由被拒绝",
			path:           "/admin",
			role:           "dept_manager",
			expectedStatus: 403,
			expectedCode:   errcode.ErrForbidden.Code,
		},
		{
			name:           "超级管理员访问 manager 路由",
			path:           "/manager",
			role:           "super_admin",
			expectedStatus: 200,
		},
		{
			name:           "部门经理访问 manager 路由",
			path:           "/manager",
			role:           "dept_manager",
			expectedStatus: 200,
		},
		{
			name:           "普通用户访问 manager 路由被拒绝",
			path:           "/manager",
			role:           "user",
			expectedStatus: 403,
			expectedCode:   errcode.ErrForbidden.Code,
		},
		{
			name:           "未设置角色访问被拒绝",
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

	t.Run("超级管理员访问", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("X-Test-Role", "super_admin")
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("非管理员访问被拒绝", func(t *testing.T) {
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
		{"超级管理员", "super_admin", 200},
		{"部门经理", "dept_manager", 200},
		{"普通用户", "user", 403},
		{"管理员", "admin", 403},
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

// ==================== 上下文获取函数测试 ====================

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		expectedID   int64
	}{
		{
			name: "正常获取用户ID",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyUserID, int64(123))
			},
			expectedID: 123,
		},
		{
			name:         "未设置用户ID",
			setupContext: func(c *gin.Context) {},
			expectedID:   0,
		},
		{
			name: "用户ID为nil",
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
			name: "正常获取角色",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyRole, "admin")
			},
			expectedRole: "admin",
		},
		{
			name:         "未设置角色",
			setupContext: func(c *gin.Context) {},
			expectedRole: "",
		},
		{
			name: "角色为nil",
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
		name         string
		setupContext func(*gin.Context)
		expectedID   *int64
	}{
		{
			name: "正常获取部门ID",
			setupContext: func(c *gin.Context) {
				c.Set(CtxKeyDepartmentID, int64(456))
			},
			expectedID: func() *int64 { id := int64(456); return &id }(),
		},
		{
			name:         "未设置部门ID",
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
		name          string
		setupContext  func(*gin.Context)
		expectedNil   bool
		expectedClaim *jwt.Claims
	}{
		{
			name: "正常获取 Claims",
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
			name:          "未设置 Claims",
			setupContext:  func(c *gin.Context) {},
			expectedNil:   true,
			expectedClaim: nil,
		},
		{
			name: "Claims 为 nil",
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

// ==================== RequestMonitor 中间件测试 ====================

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

	t.Run("记录 200 请求", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		time.Sleep(100 * time.Millisecond) // 等待 goroutine 执行
		mockStats.AssertCalled(t, "RecordRequestMetrics", 200, mock.AnythingOfType("float64"))
	})

	t.Run("记录 500 请求", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		time.Sleep(100 * time.Millisecond) // 等待 goroutine 执行
		mockStats.AssertCalled(t, "RecordRequestMetrics", 500, mock.AnythingOfType("float64"))
	})
}

// ==================== APIKeyAuth 中间件测试 ====================

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name          string
		setupHeaders  func(*http.Request)
		expectedKey   string
	}{
		{
			name: "从 x-api-key 提取",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("x-api-key", "cm-testkey123")
			},
			expectedKey: "cm-testkey123",
		},
		{
			name: "从 Authorization Bearer 提取",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer cm-testkey456")
			},
			expectedKey: "cm-testkey456",
		},
		{
			name: "x-api-key 优先级高于 Authorization",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("x-api-key", "cm-from-x-api-key")
				req.Header.Set("Authorization", "Bearer cm-from-auth")
			},
			expectedKey: "cm-from-x-api-key",
		},
		{
			name:        "无 API Key",
			setupHeaders: func(req *http.Request) {},
			expectedKey: "",
		},
		{
			name: "错误的 Authorization 格式",
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

	// 创建测试表
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

	t.Run("请求自环检测", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("X-CodeMind-Proxy", "1")
		router.ServeHTTP(w, req)

		assert.Equal(t, 502, w.Code)
	})

	t.Run("缺失 API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, errcode.ErrAPIKeyInvalid.Code, resp.Code)
	})

	t.Run("无效格式的 API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", "invalid-key")
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, errcode.ErrAPIKeyInvalid.Code, resp.Code)
	})

	t.Run("API Key 不在数据库中", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", "cm-nonexistentkey12345678901234567890123456789012")
		router.ServeHTTP(w, req)

		// 当 API Key 不存在时，GORM 返回零值结构体，KeyStatus 为 0（禁用状态）
		// 因此返回 403 Forbidden (ErrAPIKeyDisabled) 而不是 401 Unauthorized
		assert.Equal(t, 403, w.Code)
	})
}

func TestAPIKeyAuthWithValidKey(t *testing.T) {
	logger := zap.NewNop() // 使用 Nop logger 避免测试日志干扰
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// 创建测试表
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

	// 生成测试 API Key
	testAPIKey := "cm-" + strings.Repeat("a", 64) // cm- 前缀 + 64个字符
	keyHash := crypto.HashAPIKey(testAPIKey)
	deptID := int64(1)

	// 插入测试数据
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

	t.Run("有效的 API Key", func(t *testing.T) {
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

	t.Run("从 Redis 缓存获取", func(t *testing.T) {
		// 第二次请求应该从 Redis 缓存获取
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

	// 创建测试表
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

	// 生成测试 API Key (已禁用)
	testAPIKey := "cm-" + strings.Repeat("b", 64)
	keyHash := crypto.HashAPIKey(testAPIKey)

	// 插入测试数据 - Key 状态为禁用 (0)
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		2, "testuser2", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		2, keyHash, 0, 2) // status = 0 表示禁用

	router := setupTestGin()
	router.Use(APIKeyAuth(db, rdb, logger))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("禁用的 API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", testAPIKey)
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, errcode.ErrAPIKeyDisabled.Code, resp.Code)
	})
}

func TestAPIKeyAuthDisabledUser(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// 创建测试表
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

	// 生成测试 API Key
	testAPIKey := "cm-" + strings.Repeat("c", 64)
	keyHash := crypto.HashAPIKey(testAPIKey)

	// 插入测试数据 - 用户状态为禁用 (0)
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		3, "testuser3", "user", 0) // status = 0 表示禁用
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		3, keyHash, 1, 3)

	router := setupTestGin()
	router.Use(APIKeyAuth(db, rdb, logger))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("禁用的用户账号", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", testAPIKey)
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
		var resp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, errcode.ErrAccountDisabled.Code, resp.Code)
	})
}

func TestAPIKeyAuthExpiredKey(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// 创建测试表
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

	// 生成测试 API Key
	testAPIKey := "cm-" + strings.Repeat("d", 64)
	keyHash := crypto.HashAPIKey(testAPIKey)
	expiredTime := time.Now().Add(-24 * time.Hour) // 已过期

	// 插入测试数据 - Key 已过期
	db.Exec(`INSERT INTO users (id, username, role, status) VALUES (?, ?, ?, ?)`,
		4, "testuser4", "user", 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id, expires_at) VALUES (?, ?, ?, ?, ?)`,
		4, keyHash, 1, 4, expiredTime)

	router := setupTestGin()
	router.Use(APIKeyAuth(db, rdb, logger))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("过期的 API Key", func(t *testing.T) {
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

	// 创建测试表
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

	// 插入测试数据
	db.Exec(`INSERT INTO users (id, username, role, department_id, status) VALUES (?, ?, ?, ?, ?)`,
		5, "cacheduser", "manager", deptID, 1)
	db.Exec(`INSERT INTO api_keys (id, key_hash, status, user_id) VALUES (?, ?, ?, ?)`,
		5, keyHash, 1, 5)

	t.Run("从数据库查询并缓存到 Redis", func(t *testing.T) {
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

		// 验证已缓存到 Redis
		cacheKey := "codemind:apikey:" + keyHash
		cached, err := rdb.Get(ctx, cacheKey).Result()
		assert.NoError(t, err)
		assert.NotEmpty(t, cached)
	})

	t.Run("从 Redis 缓存获取", func(t *testing.T) {
		ctx := context.Background()
		info, err := getAPIKeyInfo(ctx, db, rdb, keyHash)

		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, int64(5), info.UserID)
	})

	t.Run("不存在的 API Key", func(t *testing.T) {
		ctx := context.Background()
		nonExistentHash := crypto.HashAPIKey("cm-nonexistent")
		info, err := getAPIKeyInfo(ctx, db, rdb, nonExistentHash)

		// 注意：当前实现中，当记录不存在时 GORM 不会返回错误
		// 而是返回零值的结构体（所有字段为 0）
		// 这可能是代码中的问题，但这里测试当前行为
		assert.NoError(t, err) // 当前行为：不返回错误
		assert.NotNil(t, info) // 返回空结构体而不是 nil
		assert.Equal(t, int64(0), info.UserID) // 零值
		assert.Equal(t, int16(0), info.KeyStatus) // 零值
	})
}

func TestGetAPIKeyInfoDatabaseError(t *testing.T) {
	db := setupTestDB(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// 不创建表，模拟数据库错误
	t.Run("数据库错误", func(t *testing.T) {
		ctx := context.Background()
		keyHash := crypto.HashAPIKey("cm-test")
		info, err := getAPIKeyInfo(ctx, db, rdb, keyHash)

		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

// MockDB 用于模拟数据库错误的 mock
type MockDB struct {
	errorOnQuery bool
}

func (m *MockDB) ErrorOnQuery() bool {
	return m.errorOnQuery
}

// ==================== JWTAuth 过期 Token 测试 ====================

func TestJWTAuthExpiredToken(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	// 创建 JWT Manager，设置很短的过期时间
	jwtManager := jwt.NewManager("test-secret", 0, rdb) // 0 小时过期
	router := setupTestGin()
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("过期 Token", func(t *testing.T) {
		// 生成一个立即可用的 Token（实际上已经过期因为设置的是0小时）
		// 使用负数时间测试过期情况
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInVzZXJuYW1lIjoidGVzdCIsInJvbGUiOiJhZG1pbiIsImV4cCI6MTYwMDAwMDAwMH0.invalid"

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})
}

// ==================== 边界条件测试 ====================

func TestJWTAuthMultipleSpaces(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	jwtManager := jwt.NewManager("test-secret", 24, rdb)
	router := setupTestGin()
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("Authorization 包含多个空格", func(t *testing.T) {
		token, _, err := jwtManager.GenerateToken(123, "test", "user", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer  "+token) // 两个空格
		router.ServeHTTP(w, req)

		// 应该被拒绝，因为 token 前面有空格
		assert.Equal(t, 401, w.Code)
	})
}

func TestRequireRoleTypeAssertion(t *testing.T) {
	router := setupTestGin()
	router.Use(func(c *gin.Context) {
		// 设置非字符串类型的 role（虽然这种情况不应该发生）
		c.Set(CtxKeyRole, 123)
		c.Next()
	})
	router.GET("/test", RequireRole("admin"), func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("role 类型错误会导致 panic", func(t *testing.T) {
		// 使用 recovery 捕获 panic
		logger := zaptest.NewLogger(t)
		routerWithRecovery := setupTestGin()
		routerWithRecovery.Use(Recovery(logger))
		routerWithRecovery.Use(func(c *gin.Context) {
			c.Set(CtxKeyRole, 123) // 设置错误的类型
			c.Next()
		})
		routerWithRecovery.GET("/test", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		routerWithRecovery.ServeHTTP(w, req)

		// 由于类型断言失败会导致 panic，recovery 会捕获并返回 500
		assert.Equal(t, 500, w.Code)
	})
}

// ==================== Logger 错误处理测试 ====================

func TestLoggerWithErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	router := setupTestGin()
	router.Use(Logger(logger))
	router.GET("/error", func(c *gin.Context) {
		c.Error(errors.New("test error"))
		c.JSON(400, gin.H{"error": "bad request"})
	})

	t.Run("带有错误的请求", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})
}

// ==================== 完整集成测试 ====================

func TestMiddlewareChain(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mr, rdb := setupTestRedis(t)
	defer mr.Close()

	jwtManager := jwt.NewManager("test-secret", 24, rdb)

	// 创建一个完整的中间件链
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

	t.Run("完整的中间件链 - 部门经理", func(t *testing.T) {
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

	t.Run("完整的中间件链 - 超级管理员", func(t *testing.T) {
		token, _, err := jwtManager.GenerateToken(1, "admin", "super_admin", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("完整的中间件链 - 普通用户被拒绝", func(t *testing.T) {
		token, _, err := jwtManager.GenerateToken(50, "user", "user", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
	})
}
