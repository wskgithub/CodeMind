package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codemind/internal/middleware"
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	jwtPkg "codemind/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==================== Mock Services ====================

// MockAuthService 认证服务 Mock
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(req *dto.LoginRequest, clientIP string) (*dto.LoginResponse, error) {
	args := m.Called(req, clientIP)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func (m *MockAuthService) Logout(claims *jwtPkg.Claims) error {
	args := m.Called(claims)
	return args.Error(0)
}

func (m *MockAuthService) GetProfile(userID int64) (*dto.UserDetail, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDetail), args.Error(1)
}

func (m *MockAuthService) UpdateProfile(userID int64, req *dto.UpdateProfileRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}

func (m *MockAuthService) ChangePassword(userID int64, req *dto.ChangePasswordRequest, claims *jwtPkg.Claims, clientIP string) error {
	args := m.Called(userID, req, claims, clientIP)
	return args.Error(0)
}

func (m *MockAuthService) GetLoginLockStatusByUsername(username string) (*dto.LoginLockStatusResponse, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginLockStatusResponse), args.Error(1)
}

// MockAPIKeyService API Key 服务 Mock
type MockAPIKeyService struct {
	mock.Mock
}

func (m *MockAPIKeyService) List(userID int64) ([]dto.APIKeyResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.APIKeyResponse), args.Error(1)
}

func (m *MockAPIKeyService) Create(req *dto.CreateAPIKeyRequest, userID int64, clientIP string) (*dto.APIKeyCreateResponse, error) {
	args := m.Called(req, userID, clientIP)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.APIKeyCreateResponse), args.Error(1)
}

func (m *MockAPIKeyService) UpdateStatus(keyID int64, status int16, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	args := m.Called(keyID, status, operatorID, operatorRole, operatorDeptID, clientIP)
	return args.Error(0)
}

func (m *MockAPIKeyService) Delete(keyID int64, operatorID int64, operatorRole string, clientIP string) error {
	args := m.Called(keyID, operatorID, operatorRole, clientIP)
	return args.Error(0)
}

// MockUserService 用户服务 Mock
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) List(query *dto.UserListQuery, role string, deptID *int64) ([]dto.UserDetail, int64, error) {
	args := m.Called(query, role, deptID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]dto.UserDetail), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) Create(req *dto.CreateUserRequest, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) (*dto.UserDetail, error) {
	args := m.Called(req, operatorID, operatorRole, operatorDeptID, clientIP)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDetail), args.Error(1)
}

func (m *MockUserService) GetDetail(id int64) (*dto.UserDetail, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDetail), args.Error(1)
}

func (m *MockUserService) Update(id int64, req *dto.UpdateUserRequest, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	args := m.Called(id, req, operatorID, operatorRole, operatorDeptID, clientIP)
	return args.Error(0)
}

func (m *MockUserService) Delete(id int64, operatorID int64, clientIP string) error {
	args := m.Called(id, operatorID, clientIP)
	return args.Error(0)
}

func (m *MockUserService) UpdateStatus(id int64, status int16, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	args := m.Called(id, status, operatorID, operatorRole, operatorDeptID, clientIP)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(id int64, newPassword string, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	args := m.Called(id, newPassword, operatorID, operatorRole, operatorDeptID, clientIP)
	return args.Error(0)
}

func (m *MockUserService) UnlockUser(id int64, operatorID int64, operatorRole string, operatorDeptID *int64, reason string, clientIP string) error {
	args := m.Called(id, operatorID, operatorRole, operatorDeptID, reason, clientIP)
	return args.Error(0)
}

// MockDepartmentService 部门服务 Mock
type MockDepartmentService struct {
	mock.Mock
}

func (m *MockDepartmentService) ListTree() ([]dto.DeptTree, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.DeptTree), args.Error(1)
}

func (m *MockDepartmentService) Create(req *dto.CreateDepartmentRequest, operatorID int64, clientIP string) (*model.Department, error) {
	args := m.Called(req, operatorID, clientIP)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Department), args.Error(1)
}

func (m *MockDepartmentService) GetByID(id int64) (*model.Department, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Department), args.Error(1)
}

func (m *MockDepartmentService) Update(id int64, req *dto.UpdateDepartmentRequest, operatorID int64, clientIP string) error {
	args := m.Called(id, req, operatorID, clientIP)
	return args.Error(0)
}

func (m *MockDepartmentService) Delete(id int64, operatorID int64, clientIP string) error {
	args := m.Called(id, operatorID, clientIP)
	return args.Error(0)
}

// MockLimitService 限额服务 Mock
type MockLimitService struct {
	mock.Mock
}

func (m *MockLimitService) List(query *dto.LimitListQuery) ([]model.RateLimit, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.RateLimit), args.Error(1)
}

func (m *MockLimitService) Upsert(req *dto.UpsertRateLimitRequest, operatorID int64, clientIP string) error {
	args := m.Called(req, operatorID, clientIP)
	return args.Error(0)
}

func (m *MockLimitService) Delete(id int64, operatorID int64, clientIP string) error {
	args := m.Called(id, operatorID, clientIP)
	return args.Error(0)
}

func (m *MockLimitService) GetMyLimits(userID int64, deptID *int64) (*dto.MyLimitResponse, error) {
	args := m.Called(userID, deptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MyLimitResponse), args.Error(1)
}

func (m *MockLimitService) GetLimitProgress(userID int64, deptID *int64) (*dto.LimitProgressResponse, error) {
	args := m.Called(userID, deptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LimitProgressResponse), args.Error(1)
}

// MockSystemService 系统服务 Mock
type MockSystemService struct {
	mock.Mock
}

func (m *MockSystemService) GetConfigs() ([]model.SystemConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.SystemConfig), args.Error(1)
}

func (m *MockSystemService) UpdateConfigs(req *dto.UpdateConfigsRequest, operatorID int64, clientIP string) error {
	args := m.Called(req, operatorID, clientIP)
	return args.Error(0)
}

func (m *MockSystemService) ListAnnouncements(isAdmin bool) ([]model.Announcement, error) {
	args := m.Called(isAdmin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Announcement), args.Error(1)
}

func (m *MockSystemService) CreateAnnouncement(req *dto.CreateAnnouncementRequest, authorID int64, clientIP string) (*model.Announcement, error) {
	args := m.Called(req, authorID, clientIP)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Announcement), args.Error(1)
}

func (m *MockSystemService) UpdateAnnouncement(id int64, req *dto.UpdateAnnouncementRequest, operatorID int64, clientIP string) error {
	args := m.Called(id, req, operatorID, clientIP)
	return args.Error(0)
}

func (m *MockSystemService) DeleteAnnouncement(id int64, operatorID int64, clientIP string) error {
	args := m.Called(id, operatorID, clientIP)
	return args.Error(0)
}

func (m *MockSystemService) ListAuditLogs(query *dto.AuditLogQuery) ([]model.AuditLog, int64, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.AuditLog), args.Get(1).(int64), args.Error(2)
}

// ==================== Test Helpers ====================

func setupTest() (*gin.Engine, *MockAuthService, *MockAPIKeyService, *MockUserService, *MockDepartmentService, *MockLimitService, *MockSystemService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthService := new(MockAuthService)
	mockAPIKeyService := new(MockAPIKeyService)
	mockUserService := new(MockUserService)
	mockDeptService := new(MockDepartmentService)
	mockLimitService := new(MockLimitService)
	mockSystemService := new(MockSystemService)

	return router, mockAuthService, mockAPIKeyService, mockUserService, mockDeptService, mockLimitService, mockSystemService
}

func setAuthContext(c *gin.Context, userID int64, role string, deptID *int64) {
	c.Set(middleware.CtxKeyUserID, userID)
	c.Set(middleware.CtxKeyRole, role)
	c.Set(middleware.CtxKeyUsername, "testuser")
	if deptID != nil {
		c.Set(middleware.CtxKeyDepartmentID, *deptID)
	}
	claims := &jwtPkg.Claims{
		UserID:       userID,
		Username:     "testuser",
		Role:         role,
		DepartmentID: deptID,
	}
	c.Set(middleware.CtxKeyClaims, claims)
}

func makeRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ==================== Auth Handler Tests ====================

func TestAuthHandler_Login_Success(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	expectedResp := &dto.LoginResponse{
		Token:     "test-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User: dto.UserBrief{
			ID:          1,
			Username:    "testuser",
			DisplayName: "Test User",
			Role:        "user",
		},
	}

	mockAuthService.On("Login", mock.AnythingOfType("*dto.LoginRequest"), mock.Anything).Return(expectedResp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/auth/login", dto.LoginRequest{
		Username: "testuser",
		Password: "password123",
	})

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidParams(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/auth/login", map[string]string{
		"username": "testuser",
	})

	handler.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	mockAuthService.On("Login", mock.AnythingOfType("*dto.LoginRequest"), mock.Anything).Return(nil, errcode.ErrInvalidCredentials)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/auth/login", dto.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	})

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40001), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Login_ServiceError(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	mockAuthService.On("Login", mock.AnythingOfType("*dto.LoginRequest"), mock.Anything).Return(nil, errors.New("internal error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/auth/login", dto.LoginRequest{
		Username: "testuser",
		Password: "password123",
	})

	handler.Login(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50001), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	mockAuthService.On("Logout", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/auth/logout", nil)
	setAuthContext(c, 1, "user", nil)

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Logout_Unauthorized(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/auth/logout", nil)

	handler.Logout(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40003), resp["code"])
}

func TestAuthHandler_GetProfile_Success(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	expectedProfile := &dto.UserDetail{
		ID:          1,
		Username:    "testuser",
		DisplayName: "Test User",
		Role:        "user",
		Status:      1,
	}

	mockAuthService.On("GetProfile", int64(1)).Return(expectedProfile, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/auth/profile", nil)
	setAuthContext(c, 1, "user", nil)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_GetProfile_Unauthorized(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/auth/profile", nil)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40003), resp["code"])
}

func TestAuthHandler_GetProfile_NotFound(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	mockAuthService.On("GetProfile", int64(1)).Return(nil, errcode.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/auth/profile", nil)
	setAuthContext(c, 1, "user", nil)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40306), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_UpdateProfile_Success(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	mockAuthService.On("UpdateProfile", int64(1), mock.AnythingOfType("*dto.UpdateProfileRequest")).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/auth/profile", dto.UpdateProfileRequest{
		DisplayName: strPtr("New Name"),
	})
	setAuthContext(c, 1, "user", nil)

	handler.UpdateProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_UpdateProfile_InvalidParams(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/auth/profile", "invalid json")
	setAuthContext(c, 1, "user", nil)

	handler.UpdateProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestAuthHandler_UpdateProfile_NotFound(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	mockAuthService.On("UpdateProfile", int64(1), mock.AnythingOfType("*dto.UpdateProfileRequest")).Return(errcode.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/auth/profile", dto.UpdateProfileRequest{
		DisplayName: strPtr("New Name"),
	})
	setAuthContext(c, 1, "user", nil)

	handler.UpdateProfile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40306), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	mockAuthService.On("ChangePassword", int64(1), mock.AnythingOfType("*dto.ChangePasswordRequest"), mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/auth/password", dto.ChangePasswordRequest{
		OldPassword: "oldpass123",
		NewPassword: "newpass123",
	})
	setAuthContext(c, 1, "user", nil)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_InvalidParams(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/auth/password", map[string]string{
		"old_password": "oldpass123",
	})
	setAuthContext(c, 1, "user", nil)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestAuthHandler_ChangePassword_WrongOldPassword(t *testing.T) {
	_, mockAuthService, _, _, _, _, _ := setupTest()
	handler := NewAuthHandler(mockAuthService)

	mockAuthService.On("ChangePassword", int64(1), mock.AnythingOfType("*dto.ChangePasswordRequest"), mock.Anything, mock.Anything).Return(errcode.ErrOldPasswordWrong)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/auth/password", dto.ChangePasswordRequest{
		OldPassword: "wrongpass",
		NewPassword: "newpass123",
	})
	setAuthContext(c, 1, "user", nil)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40307), resp["code"])
	mockAuthService.AssertExpectations(t)
}

// ==================== API Key Handler Tests ====================

func TestAPIKeyHandler_List_Success(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	expectedKeys := []dto.APIKeyResponse{
		{ID: 1, Name: "Key 1", KeyPrefix: "key1", Status: 1},
		{ID: 2, Name: "Key 2", KeyPrefix: "key2", Status: 1},
	}

	mockAPIKeyService.On("List", int64(1)).Return(expectedKeys, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/keys", nil)
	setAuthContext(c, 1, "user", nil)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_List_ServiceError(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	mockAPIKeyService.On("List", int64(1)).Return(nil, errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/keys", nil)
	setAuthContext(c, 1, "user", nil)

	handler.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_Create_Success(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	expectedKey := &dto.APIKeyCreateResponse{
		ID:        1,
		Name:      "New Key",
		Key:       "full-key-value",
		KeyPrefix: "prefix",
	}

	mockAPIKeyService.On("Create", mock.AnythingOfType("*dto.CreateAPIKeyRequest"), int64(1), mock.Anything).Return(expectedKey, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/keys", dto.CreateAPIKeyRequest{
		Name: "New Key",
	})
	setAuthContext(c, 1, "user", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_Create_InvalidParams(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/keys", map[string]string{})
	setAuthContext(c, 1, "user", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestAPIKeyHandler_Create_KeyLimit(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	mockAPIKeyService.On("Create", mock.AnythingOfType("*dto.CreateAPIKeyRequest"), int64(1), mock.Anything).Return(nil, errcode.ErrAPIKeyLimit)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/keys", dto.CreateAPIKeyRequest{
		Name: "New Key",
	})
	setAuthContext(c, 1, "user", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40304), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_UpdateStatus_Success(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	mockAPIKeyService.On("UpdateStatus", int64(1), int16(0), int64(1), "user", (*int64)(nil), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/keys/1/status", dto.UpdateStatusRequest{Status: 0})
	setAuthContext(c, 1, "user", nil)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_UpdateStatus_InvalidID(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("PUT", "/api/v1/keys/invalid/status", dto.UpdateStatusRequest{Status: 0})
	setAuthContext(c, 1, "user", nil)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestAPIKeyHandler_UpdateStatus_InvalidParams(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/keys/1/status", "invalid")
	setAuthContext(c, 1, "user", nil)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestAPIKeyHandler_UpdateStatus_Forbidden(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	mockAPIKeyService.On("UpdateStatus", int64(1), int16(0), int64(2), "user", (*int64)(nil), mock.Anything).Return(errcode.ErrForbidden)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/keys/1/status", dto.UpdateStatusRequest{Status: 0})
	setAuthContext(c, 2, "user", nil)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40101), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_UpdateStatus_NotFound(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	mockAPIKeyService.On("UpdateStatus", int64(999), int16(0), int64(1), "user", (*int64)(nil), mock.Anything).Return(errcode.ErrAPIKeyNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("PUT", "/api/v1/keys/999/status", dto.UpdateStatusRequest{Status: 0})
	setAuthContext(c, 1, "user", nil)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40308), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_Delete_Success(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	mockAPIKeyService.On("Delete", int64(1), int64(1), "user", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("DELETE", "/api/v1/keys/1", nil)
	setAuthContext(c, 1, "user", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_Delete_InvalidID(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("DELETE", "/api/v1/keys/invalid", nil)
	setAuthContext(c, 1, "user", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestAPIKeyHandler_Delete_Forbidden(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	mockAPIKeyService.On("Delete", int64(1), int64(2), "user", mock.Anything).Return(errcode.ErrForbidden)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("DELETE", "/api/v1/keys/1", nil)
	setAuthContext(c, 2, "user", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40101), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

func TestAPIKeyHandler_Delete_NotFound(t *testing.T) {
	_, _, mockAPIKeyService, _, _, _, _ := setupTest()
	handler := NewAPIKeyHandler(mockAPIKeyService)

	mockAPIKeyService.On("Delete", int64(999), int64(1), "user", mock.Anything).Return(errcode.ErrAPIKeyNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("DELETE", "/api/v1/keys/999", nil)
	setAuthContext(c, 1, "user", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40308), resp["code"])
	mockAPIKeyService.AssertExpectations(t)
}

// ==================== User Handler Tests ====================

func TestUserHandler_List_Success(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	expectedUsers := []dto.UserDetail{
		{ID: 1, Username: "user1", DisplayName: "User 1", Role: "user"},
		{ID: 2, Username: "user2", DisplayName: "User 2", Role: "user"},
	}

	mockUserService.On("List", mock.AnythingOfType("*dto.UserListQuery"), "super_admin", (*int64)(nil)).Return(expectedUsers, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/users?page=1&page_size=10", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_List_InvalidParams(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/users?status=invalid", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.List(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_List_ServiceError(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("List", mock.AnythingOfType("*dto.UserListQuery"), "super_admin", (*int64)(nil)).Return(nil, int64(0), errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/users", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_Create_Success(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	expectedUser := &dto.UserDetail{
		ID:          1,
		Username:    "newuser",
		DisplayName: "New User",
		Role:        "user",
	}

	mockUserService.On("Create", mock.AnythingOfType("*dto.CreateUserRequest"), int64(1), "super_admin", (*int64)(nil), mock.Anything).Return(expectedUser, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/users", dto.CreateUserRequest{
		Username:    "newuser",
		Password:    "password123",
		DisplayName: "New User",
		Role:        "user",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_Create_InvalidParams(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/users", map[string]string{
		"username": "newuser",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_Create_UsernameExists(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("Create", mock.AnythingOfType("*dto.CreateUserRequest"), int64(1), "super_admin", (*int64)(nil), mock.Anything).Return(nil, errcode.ErrUsernameExists)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/users", dto.CreateUserRequest{
		Username:    "existinguser",
		Password:    "password123",
		DisplayName: "Existing User",
		Role:        "user",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40301), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_Create_Forbidden(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("Create", mock.AnythingOfType("*dto.CreateUserRequest"), int64(2), "dept_manager", mock.Anything, mock.Anything).Return(nil, errcode.ErrForbiddenUser)

	deptID := int64(1)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/users", dto.CreateUserRequest{
		Username:    "newuser",
		Password:    "password123",
		DisplayName: "New User",
		Role:        "super_admin",
	})
	setAuthContext(c, 2, "dept_manager", &deptID)

	handler.Create(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40102), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetDetail_Success(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	expectedUser := &dto.UserDetail{
		ID:          1,
		Username:    "testuser",
		DisplayName: "Test User",
		Role:        "user",
	}

	mockUserService.On("GetDetail", int64(1)).Return(expectedUser, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("GET", "/api/v1/users/1", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.GetDetail(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetDetail_InvalidID(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("GET", "/api/v1/users/invalid", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.GetDetail(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_GetDetail_NotFound(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("GetDetail", int64(999)).Return(nil, errcode.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("GET", "/api/v1/users/999", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.GetDetail(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40306), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_Update_Success(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("Update", int64(1), mock.AnythingOfType("*dto.UpdateUserRequest"), int64(1), "super_admin", (*int64)(nil), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/users/1", dto.UpdateUserRequest{
		DisplayName: strPtr("Updated Name"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("PUT", "/api/v1/users/invalid", dto.UpdateUserRequest{
		DisplayName: strPtr("Updated Name"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_Update_InvalidParams(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/users/1", "invalid")
	setAuthContext(c, 1, "super_admin", nil)

	handler.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("Update", int64(999), mock.AnythingOfType("*dto.UpdateUserRequest"), int64(1), "super_admin", (*int64)(nil), mock.Anything).Return(errcode.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("PUT", "/api/v1/users/999", dto.UpdateUserRequest{
		DisplayName: strPtr("Updated Name"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40306), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_Update_Forbidden(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("Update", int64(1), mock.AnythingOfType("*dto.UpdateUserRequest"), int64(2), "dept_manager", mock.Anything, mock.Anything).Return(errcode.ErrForbiddenUser)

	deptID := int64(1)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/users/1", dto.UpdateUserRequest{
		DisplayName: strPtr("Updated Name"),
	})
	setAuthContext(c, 2, "dept_manager", &deptID)

	handler.Update(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40102), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_Delete_Success(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("Delete", int64(2), int64(1), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = makeRequest("DELETE", "/api/v1/users/2", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_Delete_InvalidID(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("DELETE", "/api/v1/users/invalid", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("Delete", int64(999), int64(1), mock.Anything).Return(errcode.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("DELETE", "/api/v1/users/999", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40306), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_UpdateStatus_Success(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("UpdateStatus", int64(2), int16(0), int64(1), "super_admin", (*int64)(nil), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = makeRequest("PUT", "/api/v1/users/2/status", dto.UpdateStatusRequest{Status: 0})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_UpdateStatus_InvalidID(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("PUT", "/api/v1/users/invalid/status", dto.UpdateStatusRequest{Status: 0})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_UpdateStatus_NotFound(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("UpdateStatus", int64(999), int16(0), int64(1), "super_admin", (*int64)(nil), mock.Anything).Return(errcode.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("PUT", "/api/v1/users/999/status", dto.UpdateStatusRequest{Status: 0})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40306), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_UpdateStatus_Forbidden(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("UpdateStatus", int64(1), int16(0), int64(2), "dept_manager", mock.Anything, mock.Anything).Return(errcode.ErrForbiddenUser)

	deptID := int64(1)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/users/1/status", dto.UpdateStatusRequest{Status: 0})
	setAuthContext(c, 2, "dept_manager", &deptID)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40102), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_ResetPassword_Success(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("ResetPassword", int64(2), "newpassword123", int64(1), "super_admin", (*int64)(nil), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = makeRequest("PUT", "/api/v1/users/2/reset-password", dto.ResetPasswordRequest{
		NewPassword: "newpassword123",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_ResetPassword_InvalidID(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("PUT", "/api/v1/users/invalid/reset-password", dto.ResetPasswordRequest{
		NewPassword: "newpassword123",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_ResetPassword_InvalidParams(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = makeRequest("PUT", "/api/v1/users/2/reset-password", map[string]string{})
	setAuthContext(c, 1, "super_admin", nil)

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_ResetPassword_NotFound(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("ResetPassword", int64(999), "newpassword123", int64(1), "super_admin", (*int64)(nil), mock.Anything).Return(errcode.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("PUT", "/api/v1/users/999/reset-password", dto.ResetPasswordRequest{
		NewPassword: "newpassword123",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40306), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_UnlockUser_Success(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("UnlockUser", int64(2), int64(1), "super_admin", (*int64)(nil), "Test reason", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = makeRequest("PUT", "/api/v1/users/2/unlock", dto.UnlockUserRequest{
		Reason: "Test reason",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UnlockUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_UnlockUser_InvalidID(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("PUT", "/api/v1/users/invalid/unlock", dto.UnlockUserRequest{})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UnlockUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestUserHandler_UnlockUser_NotFound(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("UnlockUser", int64(999), int64(1), "super_admin", (*int64)(nil), "", mock.Anything).Return(errcode.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("PUT", "/api/v1/users/999/unlock", dto.UnlockUserRequest{})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UnlockUser(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40306), resp["code"])
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_UnlockUser_Forbidden(t *testing.T) {
	_, _, _, mockUserService, _, _, _ := setupTest()
	handler := NewUserHandler(mockUserService)

	mockUserService.On("UnlockUser", int64(1), int64(2), "dept_manager", mock.Anything, "", mock.Anything).Return(errcode.ErrForbiddenUser)

	deptID := int64(1)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/users/1/unlock", dto.UnlockUserRequest{})
	setAuthContext(c, 2, "dept_manager", &deptID)

	handler.UnlockUser(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40102), resp["code"])
	mockUserService.AssertExpectations(t)
}

// ==================== Department Handler Tests ====================

func TestDepartmentHandler_List_Success(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	expectedTree := []dto.DeptTree{
		{ID: 1, Name: "Dept 1", UserCount: 5},
		{ID: 2, Name: "Dept 2", UserCount: 3},
	}

	mockDeptService.On("ListTree").Return(expectedTree, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/departments", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_List_ServiceError(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	mockDeptService.On("ListTree").Return(nil, errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/departments", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_Create_Success(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	expectedDept := &model.Department{
		ID:   1,
		Name: "New Department",
	}

	mockDeptService.On("Create", mock.AnythingOfType("*dto.CreateDepartmentRequest"), int64(1), mock.Anything).Return(expectedDept, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/departments", dto.CreateDepartmentRequest{
		Name: "New Department",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_Create_InvalidParams(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/departments", map[string]string{})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestDepartmentHandler_Create_DuplicateName(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	mockDeptService.On("Create", mock.AnythingOfType("*dto.CreateDepartmentRequest"), int64(1), mock.Anything).Return(nil, errcode.ErrInvalidParams.WithMessage("部门名称已存在"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/departments", dto.CreateDepartmentRequest{
		Name: "Existing Department",
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_GetDetail_Success(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	expectedDept := &model.Department{
		ID:   1,
		Name: "Department 1",
	}

	mockDeptService.On("GetByID", int64(1)).Return(expectedDept, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("GET", "/api/v1/departments/1", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.GetDetail(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_GetDetail_InvalidID(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("GET", "/api/v1/departments/invalid", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.GetDetail(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestDepartmentHandler_GetDetail_NotFound(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	mockDeptService.On("GetByID", int64(999)).Return(nil, errcode.ErrDeptNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("GET", "/api/v1/departments/999", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.GetDetail(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40303), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_Update_Success(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	mockDeptService.On("Update", int64(1), mock.AnythingOfType("*dto.UpdateDepartmentRequest"), int64(1), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/departments/1", dto.UpdateDepartmentRequest{
		Name: strPtr("Updated Department"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_Update_InvalidID(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("PUT", "/api/v1/departments/invalid", dto.UpdateDepartmentRequest{
		Name: strPtr("Updated Department"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestDepartmentHandler_Update_InvalidParams(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/departments/1", "invalid")
	setAuthContext(c, 1, "super_admin", nil)

	handler.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestDepartmentHandler_Update_NotFound(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	mockDeptService.On("Update", int64(999), mock.AnythingOfType("*dto.UpdateDepartmentRequest"), int64(1), mock.Anything).Return(errcode.ErrDeptNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("PUT", "/api/v1/departments/999", dto.UpdateDepartmentRequest{
		Name: strPtr("Updated Department"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40303), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_Delete_Success(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	mockDeptService.On("Delete", int64(1), int64(1), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("DELETE", "/api/v1/departments/1", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_Delete_InvalidID(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("DELETE", "/api/v1/departments/invalid", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestDepartmentHandler_Delete_NotFound(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	mockDeptService.On("Delete", int64(999), int64(1), mock.Anything).Return(errcode.ErrDeptNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("DELETE", "/api/v1/departments/999", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40303), resp["code"])
	mockDeptService.AssertExpectations(t)
}

func TestDepartmentHandler_Delete_HasUsers(t *testing.T) {
	_, _, _, _, mockDeptService, _, _ := setupTest()
	handler := NewDepartmentHandler(mockDeptService)

	mockDeptService.On("Delete", int64(1), int64(1), mock.Anything).Return(errcode.ErrDeptHasUsers)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("DELETE", "/api/v1/departments/1", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40305), resp["code"])
	mockDeptService.AssertExpectations(t)
}

// ==================== Limit Handler Tests ====================

func TestLimitHandler_List_Success(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	expectedLimits := []model.RateLimit{
		{ID: 1, TargetType: "user", TargetID: 1, MaxTokens: 1000},
		{ID: 2, TargetType: "department", TargetID: 1, MaxTokens: 5000},
	}

	mockLimitService.On("List", mock.AnythingOfType("*dto.LimitListQuery")).Return(expectedLimits, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/limits", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockLimitService.AssertExpectations(t)
}

func TestLimitHandler_List_InvalidParams(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/limits?target_type=invalid", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.List(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestLimitHandler_List_ServiceError(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	mockLimitService.On("List", mock.AnythingOfType("*dto.LimitListQuery")).Return(nil, errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/limits", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockLimitService.AssertExpectations(t)
}

func TestLimitHandler_Upsert_Success(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	mockLimitService.On("Upsert", mock.AnythingOfType("*dto.UpsertRateLimitRequest"), int64(1), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/limits", dto.UpsertRateLimitRequest{
		TargetType: "user",
		TargetID:   1,
		Period:     "daily",
		MaxTokens:  1000,
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Upsert(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockLimitService.AssertExpectations(t)
}

func TestLimitHandler_Upsert_InvalidParams(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/limits", map[string]string{})
	setAuthContext(c, 1, "super_admin", nil)

	handler.Upsert(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestLimitHandler_Upsert_Forbidden(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/limits", dto.UpsertRateLimitRequest{
		TargetType: "department",
		TargetID:   1,
		Period:     "daily",
		MaxTokens:  1000,
	})
	setAuthContext(c, 2, "dept_manager", nil)

	handler.Upsert(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40101), resp["code"])
}

func TestLimitHandler_Upsert_DeptManagerSuccess(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	mockLimitService.On("Upsert", mock.AnythingOfType("*dto.UpsertRateLimitRequest"), int64(2), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/limits", dto.UpsertRateLimitRequest{
		TargetType: "user",
		TargetID:   3,
		Period:     "daily",
		MaxTokens:  1000,
	})
	setAuthContext(c, 2, "dept_manager", nil)

	handler.Upsert(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockLimitService.AssertExpectations(t)
}

func TestLimitHandler_GetMyLimits_Success(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	expectedLimits := &dto.MyLimitResponse{
		Limits: map[string]dto.LimitDetail{
			"daily": {MaxTokens: 1000, UsedTokens: 100, RemainingTokens: 900, UsagePercent: 10},
		},
		Concurrency: dto.ConcurrencyInfo{Max: 5, Current: 1},
	}

	mockLimitService.On("GetMyLimits", int64(1), (*int64)(nil)).Return(expectedLimits, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/limits/my", nil)
	setAuthContext(c, 1, "user", nil)

	handler.GetMyLimits(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockLimitService.AssertExpectations(t)
}

func TestLimitHandler_GetMyLimits_ServiceError(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	mockLimitService.On("GetMyLimits", int64(1), (*int64)(nil)).Return(nil, errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/limits/my", nil)
	setAuthContext(c, 1, "user", nil)

	handler.GetMyLimits(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockLimitService.AssertExpectations(t)
}

func TestLimitHandler_GetMyProgress_Success(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	expectedProgress := &dto.LimitProgressResponse{
		Limits: []dto.LimitProgressItem{
			{RuleID: 1, Period: "daily", MaxTokens: 1000, UsedTokens: 100},
		},
		Concurrency: dto.ConcurrencyInfo{Max: 5, Current: 1},
	}

	mockLimitService.On("GetLimitProgress", int64(1), (*int64)(nil)).Return(expectedProgress, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/limits/my/progress", nil)
	setAuthContext(c, 1, "user", nil)

	handler.GetMyProgress(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockLimitService.AssertExpectations(t)
}

func TestLimitHandler_Delete_Success(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	mockLimitService.On("Delete", int64(1), int64(1), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("DELETE", "/api/v1/limits/1", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockLimitService.AssertExpectations(t)
}

func TestLimitHandler_Delete_InvalidID(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("DELETE", "/api/v1/limits/invalid", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestLimitHandler_Delete_NotFound(t *testing.T) {
	_, _, _, _, _, mockLimitService, _ := setupTest()
	handler := NewLimitHandler(mockLimitService)

	mockLimitService.On("Delete", int64(999), int64(1), mock.Anything).Return(errcode.ErrRecordNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("DELETE", "/api/v1/limits/999", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40309), resp["code"])
	mockLimitService.AssertExpectations(t)
}

// ==================== System Handler Tests ====================

func TestSystemHandler_GetConfigs_Success(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	expectedConfigs := []model.SystemConfig{
		{ConfigKey: "site_name", ConfigValue: "CodeMind"},
		{ConfigKey: "max_keys", ConfigValue: "5"},
	}

	mockSystemService.On("GetConfigs").Return(expectedConfigs, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/system/configs", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.GetConfigs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_GetConfigs_ServiceError(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("GetConfigs").Return(nil, errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/system/configs", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.GetConfigs(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_UpdateConfigs_Success(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("UpdateConfigs", mock.AnythingOfType("*dto.UpdateConfigsRequest"), int64(1), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/system/configs", dto.UpdateConfigsRequest{
		Configs: []dto.ConfigItem{
			{Key: "site_name", Value: "New Name"},
		},
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateConfigs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_UpdateConfigs_InvalidParams(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/system/configs", map[string]string{})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateConfigs(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestSystemHandler_UpdateConfigs_ServiceError(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("UpdateConfigs", mock.AnythingOfType("*dto.UpdateConfigsRequest"), int64(1), mock.Anything).Return(errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("PUT", "/api/v1/system/configs", dto.UpdateConfigsRequest{
		Configs: []dto.ConfigItem{
			{Key: "site_name", Value: "New Name"},
		},
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateConfigs(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_ListAnnouncements_Success(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	expectedAnns := []model.Announcement{
		{ID: 1, Title: "Announcement 1", Content: "Content 1"},
		{ID: 2, Title: "Announcement 2", Content: "Content 2"},
	}

	mockSystemService.On("ListAnnouncements", true).Return(expectedAnns, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/system/announcements", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.ListAnnouncements(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_ListAnnouncements_AsUser(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	expectedAnns := []model.Announcement{
		{ID: 1, Title: "Announcement 1", Content: "Content 1", Status: 1},
	}

	mockSystemService.On("ListAnnouncements", false).Return(expectedAnns, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/system/announcements", nil)
	setAuthContext(c, 1, "user", nil)

	handler.ListAnnouncements(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_ListAnnouncements_ServiceError(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("ListAnnouncements", true).Return(nil, errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/system/announcements", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.ListAnnouncements(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_CreateAnnouncement_Success(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	expectedAnn := &model.Announcement{
		ID:      1,
		Title:   "New Announcement",
		Content: "Content",
	}

	mockSystemService.On("CreateAnnouncement", mock.AnythingOfType("*dto.CreateAnnouncementRequest"), int64(1), mock.Anything).Return(expectedAnn, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/system/announcements", dto.CreateAnnouncementRequest{
		Title:   "New Announcement",
		Content: "Content",
		Status:  1,
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.CreateAnnouncement(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_CreateAnnouncement_InvalidParams(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/system/announcements", map[string]string{})
	setAuthContext(c, 1, "super_admin", nil)

	handler.CreateAnnouncement(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestSystemHandler_CreateAnnouncement_ServiceError(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("CreateAnnouncement", mock.AnythingOfType("*dto.CreateAnnouncementRequest"), int64(1), mock.Anything).Return(nil, errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("POST", "/api/v1/system/announcements", dto.CreateAnnouncementRequest{
		Title:   "New Announcement",
		Content: "Content",
		Status:  1,
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.CreateAnnouncement(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_UpdateAnnouncement_Success(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("UpdateAnnouncement", int64(1), mock.AnythingOfType("*dto.UpdateAnnouncementRequest"), int64(1), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/system/announcements/1", dto.UpdateAnnouncementRequest{
		Title: strPtr("Updated Title"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateAnnouncement(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_UpdateAnnouncement_InvalidID(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("PUT", "/api/v1/system/announcements/invalid", dto.UpdateAnnouncementRequest{
		Title: strPtr("Updated Title"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateAnnouncement(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestSystemHandler_UpdateAnnouncement_InvalidParams(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("PUT", "/api/v1/system/announcements/1", "invalid")
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateAnnouncement(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestSystemHandler_UpdateAnnouncement_NotFound(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("UpdateAnnouncement", int64(999), mock.AnythingOfType("*dto.UpdateAnnouncementRequest"), int64(1), mock.Anything).Return(errcode.ErrRecordNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("PUT", "/api/v1/system/announcements/999", dto.UpdateAnnouncementRequest{
		Title: strPtr("Updated Title"),
	})
	setAuthContext(c, 1, "super_admin", nil)

	handler.UpdateAnnouncement(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40309), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_DeleteAnnouncement_Success(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("DeleteAnnouncement", int64(1), int64(1), mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = makeRequest("DELETE", "/api/v1/system/announcements/1", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.DeleteAnnouncement(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_DeleteAnnouncement_InvalidID(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = makeRequest("DELETE", "/api/v1/system/announcements/invalid", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.DeleteAnnouncement(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestSystemHandler_DeleteAnnouncement_NotFound(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("DeleteAnnouncement", int64(999), int64(1), mock.Anything).Return(errcode.ErrRecordNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = makeRequest("DELETE", "/api/v1/system/announcements/999", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.DeleteAnnouncement(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40309), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_ListAuditLogs_Success(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	expectedLogs := []model.AuditLog{
		{ID: 1, OperatorID: 1, Action: "login"},
		{ID: 2, OperatorID: 2, Action: "logout"},
	}

	mockSystemService.On("ListAuditLogs", mock.AnythingOfType("*dto.AuditLogQuery")).Return(expectedLogs, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/system/audit-logs?page=1&page_size=10", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	mockSystemService.AssertExpectations(t)
}

func TestSystemHandler_ListAuditLogs_InvalidParams(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// page_size=-1 violates min=1 validation
	c.Request = makeRequest("GET", "/api/v1/system/audit-logs?page_size=-1", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40201), resp["code"])
}

func TestSystemHandler_ListAuditLogs_ServiceError(t *testing.T) {
	_, _, _, _, _, _, mockSystemService := setupTest()
	handler := NewSystemHandler(mockSystemService)

	mockSystemService.On("ListAuditLogs", mock.AnythingOfType("*dto.AuditLogQuery")).Return(nil, int64(0), errcode.ErrDatabase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = makeRequest("GET", "/api/v1/system/audit-logs", nil)
	setAuthContext(c, 1, "super_admin", nil)

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50003), resp["code"])
	mockSystemService.AssertExpectations(t)
}

// ==================== Helper Functions ====================

func strPtr(s string) *string {
	return &s
}
