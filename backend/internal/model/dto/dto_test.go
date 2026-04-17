package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== UserListQuery tests ====================

func TestUserListQuery_GetPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected int
	}{
		{"default page - 0", 0, 1},
		{"default page - negative", -1, 1},
		{"normal page - 1", 1, 1},
		{"normal page - 10", 10, 10},
		{"large page - 999", 999, 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &UserListQuery{Page: tt.page}
			assert.Equal(t, tt.expected, q.GetPage())
		})
	}
}

func TestUserListQuery_GetPageSize(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		expected int
	}{
		{"default size - 0", 0, 20},
		{"default size - negative", -1, 20},
		{"normal size - 1", 1, 1},
		{"normal size - 50", 50, 50},
		{"max size - 100", 100, 100},
		{"exceeds max - 101", 101, 100},
		{"exceeds max - 200", 200, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &UserListQuery{PageSize: tt.pageSize}
			assert.Equal(t, tt.expected, q.GetPageSize())
		})
	}
}

// ==================== RankingQuery tests ====================

func TestRankingQuery_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		expected int
	}{
		{"default limit - 0", 0, 10},
		{"default limit - negative", -1, 10},
		{"default limit - -10", -10, 10},
		{"normal limit - 1", 1, 1},
		{"normal limit - 50", 50, 50},
		{"max limit - 100", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &RankingQuery{Limit: tt.limit}
			assert.Equal(t, tt.expected, q.GetLimit())
		})
	}
}

// ==================== AuditLogQuery tests ====================

func TestAuditLogQuery_GetPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected int
	}{
		{"default page - 0", 0, 1},
		{"default page - negative", -1, 1},
		{"normal page - 1", 1, 1},
		{"normal page - 5", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &AuditLogQuery{Page: tt.page}
			assert.Equal(t, tt.expected, q.GetPage())
		})
	}
}

func TestAuditLogQuery_GetPageSize(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		expected int
	}{
		{"default size - 0", 0, 20},
		{"default size - negative", -5, 20},
		{"normal size - 1", 1, 1},
		{"normal size - 50", 50, 50},
		{"max size - 100", 100, 100},
		{"exceeds max - 101", 101, 100},
		{"exceeds max - 1000", 1000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &AuditLogQuery{PageSize: tt.pageSize}
			assert.Equal(t, tt.expected, q.GetPageSize())
		})
	}
}

// ==================== Struct field validation tests ====================

func TestLoginRequest_StructFields(t *testing.T) {
	r := &LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	assert.Equal(t, "testuser", r.Username)
	assert.Equal(t, "password123", r.Password)
}

func TestChangePasswordRequest_StructFields(t *testing.T) {
	r := &ChangePasswordRequest{
		OldPassword: "old123",
		NewPassword: "new123",
	}
	assert.Equal(t, "old123", r.OldPassword)
	assert.Equal(t, "new123", r.NewPassword)
}

func TestUpdateProfileRequest_StructFields(t *testing.T) {
	displayName := "Test User"
	email := "test@example.com"
	phone := "1234567890"
	r := &UpdateProfileRequest{
		DisplayName: &displayName,
		Email:       &email,
		Phone:       &phone,
	}
	assert.Equal(t, &displayName, r.DisplayName)
	assert.Equal(t, &email, r.Email)
	assert.Equal(t, &phone, r.Phone)
}

func TestCreateUserRequest_StructFields(t *testing.T) {
	deptID := int64(1)
	r := &CreateUserRequest{
		Username:     "testuser",
		Password:     "password123",
		DisplayName:  "Test User",
		Email:        "test@example.com",
		Phone:        "1234567890",
		Role:         "user",
		DepartmentID: &deptID,
	}
	assert.Equal(t, "testuser", r.Username)
	assert.Equal(t, "password123", r.Password)
	assert.Equal(t, "Test User", r.DisplayName)
	assert.Equal(t, "test@example.com", r.Email)
	assert.Equal(t, "user", r.Role)
	assert.Equal(t, &deptID, r.DepartmentID)
}

func TestUpdateUserRequest_StructFields(t *testing.T) {
	displayName := "Updated User"
	email := "updated@example.com"
	role := "dept_manager"
	status := int16(1)
	deptID := int64(2)
	r := &UpdateUserRequest{
		DisplayName:  &displayName,
		Email:        &email,
		Role:         &role,
		DepartmentID: &deptID,
		Status:       &status,
	}
	assert.Equal(t, &displayName, r.DisplayName)
	assert.Equal(t, &email, r.Email)
	assert.Equal(t, &role, r.Role)
	assert.Equal(t, &status, r.Status)
	assert.Equal(t, &deptID, r.DepartmentID)
}

func TestUpdateStatusRequest_StructFields(t *testing.T) {
	r := &UpdateStatusRequest{Status: 1}
	assert.Equal(t, int16(1), r.Status)
}

func TestResetPasswordRequest_StructFields(t *testing.T) {
	r := &ResetPasswordRequest{NewPassword: "newpassword123"}
	assert.Equal(t, "newpassword123", r.NewPassword)
}

func TestUnlockUserRequest_StructFields(t *testing.T) {
	r := &UnlockUserRequest{Reason: "Test unlock"}
	assert.Equal(t, "Test unlock", r.Reason)
}

func TestUserListQuery_StructFields(t *testing.T) {
	deptID := int64(1)
	status := int16(1)
	r := &UserListQuery{
		Page:         1,
		PageSize:     20,
		Keyword:      "test",
		DepartmentID: &deptID,
		Role:         "user",
		Status:       &status,
	}
	assert.Equal(t, 1, r.Page)
	assert.Equal(t, 20, r.PageSize)
	assert.Equal(t, "test", r.Keyword)
	assert.Equal(t, &deptID, r.DepartmentID)
	assert.Equal(t, "user", r.Role)
	assert.Equal(t, &status, r.Status)
}

func TestCreateDepartmentRequest_StructFields(t *testing.T) {
	parentID := int64(1)
	managerID := int64(2)
	description := "Test Dept"
	r := &CreateDepartmentRequest{
		Name:        "Test",
		Description: &description,
		ParentID:    &parentID,
		ManagerID:   &managerID,
	}
	assert.Equal(t, "Test", r.Name)
	assert.Equal(t, &description, r.Description)
	assert.Equal(t, &parentID, r.ParentID)
	assert.Equal(t, &managerID, r.ManagerID)
}

func TestUpdateDepartmentRequest_StructFields(t *testing.T) {
	name := "Updated"
	description := "Updated Desc"
	r := &UpdateDepartmentRequest{
		Name:        &name,
		Description: &description,
	}
	assert.Equal(t, &name, r.Name)
	assert.Equal(t, &description, r.Description)
}

func TestCreateAPIKeyRequest_StructFields(t *testing.T) {
	r := &CreateAPIKeyRequest{Name: "Test Key"}
	assert.Equal(t, "Test Key", r.Name)
}

func TestStatsQuery_StructFields(t *testing.T) {
	userID := int64(1)
	deptID := int64(2)
	r := &StatsQuery{
		Period:       "daily",
		StartDate:    "2024-01-01",
		EndDate:      "2024-01-31",
		UserID:       &userID,
		DepartmentID: &deptID,
	}
	assert.Equal(t, "daily", r.Period)
	assert.Equal(t, "2024-01-01", r.StartDate)
	assert.Equal(t, "2024-01-31", r.EndDate)
	assert.Equal(t, &userID, r.UserID)
	assert.Equal(t, &deptID, r.DepartmentID)
}

func TestRankingQuery_StructFields(t *testing.T) {
	r := &RankingQuery{
		Type:   "user",
		Period: "weekly",
		Limit:  20,
	}
	assert.Equal(t, "user", r.Type)
	assert.Equal(t, "weekly", r.Period)
	assert.Equal(t, 20, r.Limit)
}

func TestKeyUsageQuery_StructFields(t *testing.T) {
	r := &KeyUsageQuery{
		Period:    "monthly",
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
	}
	assert.Equal(t, "monthly", r.Period)
	assert.Equal(t, "2024-01-01", r.StartDate)
	assert.Equal(t, "2024-01-31", r.EndDate)
}

func TestUpsertRateLimitRequest_StructFields(t *testing.T) {
	r := &UpsertRateLimitRequest{
		TargetType:     "user",
		TargetID:       1,
		Period:         "daily",
		PeriodHours:    24,
		MaxTokens:      100000,
		MaxRequests:    1000,
		MaxConcurrency: 5,
		AlertThreshold: 80,
	}
	assert.Equal(t, "user", r.TargetType)
	assert.Equal(t, int64(1), r.TargetID)
	assert.Equal(t, "daily", r.Period)
	assert.Equal(t, 24, r.PeriodHours)
	assert.Equal(t, int64(100000), r.MaxTokens)
	assert.Equal(t, 1000, r.MaxRequests)
	assert.Equal(t, 5, r.MaxConcurrency)
	assert.Equal(t, int16(80), r.AlertThreshold)
}

func TestLimitListQuery_StructFields(t *testing.T) {
	targetID := int64(1)
	r := &LimitListQuery{
		TargetType: "department",
		TargetID:   &targetID,
	}
	assert.Equal(t, "department", r.TargetType)
	assert.Equal(t, &targetID, r.TargetID)
}

func TestCreateLLMBackendRequest_StructFields(t *testing.T) {
	r := &CreateLLMBackendRequest{
		Name:                 "test-backend",
		DisplayName:          "Test Backend",
		BaseURL:              "https://api.test.com",
		APIKey:               "sk-test",
		Format:               "openai",
		Weight:               100,
		MaxConcurrency:       50,
		HealthCheckURL:       "https://api.test.com/health",
		TimeoutSeconds:       300,
		StreamTimeoutSeconds: 600,
		ModelPatterns:        "gpt-*",
	}
	assert.Equal(t, "test-backend", r.Name)
	assert.Equal(t, "Test Backend", r.DisplayName)
	assert.Equal(t, "https://api.test.com", r.BaseURL)
	assert.Equal(t, "sk-test", r.APIKey)
	assert.Equal(t, "openai", r.Format)
	assert.Equal(t, 100, r.Weight)
}

func TestUpdateLLMBackendRequest_StructFields(t *testing.T) {
	name := "Updated"
	weight := 200
	status := int16(1)
	r := &UpdateLLMBackendRequest{
		DisplayName: &name,
		Weight:      &weight,
		Status:      &status,
	}
	assert.Equal(t, &name, r.DisplayName)
	assert.Equal(t, &weight, r.Weight)
	assert.Equal(t, &status, r.Status)
}

func TestUpdateConfigsRequest_StructFields(t *testing.T) {
	r := &UpdateConfigsRequest{
		Configs: []ConfigItem{
			{Key: "test.key", Value: "test_value"},
		},
	}
	assert.Len(t, r.Configs, 1)
	assert.Equal(t, "test.key", r.Configs[0].Key)
	assert.Equal(t, "test_value", r.Configs[0].Value)
}

func TestConfigItem_StructFields(t *testing.T) {
	item := ConfigItem{Key: "key", Value: "value"}
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, "value", item.Value)
}

func TestCreateAnnouncementRequest_StructFields(t *testing.T) {
	r := &CreateAnnouncementRequest{
		Title:   "Test Title",
		Content: "Test Content",
		Pinned:  true,
		Status:  1,
	}
	assert.Equal(t, "Test Title", r.Title)
	assert.Equal(t, "Test Content", r.Content)
	assert.True(t, r.Pinned)
	assert.Equal(t, int16(1), r.Status)
}

func TestUpdateAnnouncementRequest_StructFields(t *testing.T) {
	title := "Updated"
	content := "Updated Content"
	pinned := false
	status := int16(0)
	r := &UpdateAnnouncementRequest{
		Title:   &title,
		Content: &content,
		Pinned:  &pinned,
		Status:  &status,
	}
	assert.Equal(t, &title, r.Title)
	assert.Equal(t, &content, r.Content)
	assert.Equal(t, &pinned, r.Pinned)
	assert.Equal(t, &status, r.Status)
}

func TestAuditLogQuery_StructFields(t *testing.T) {
	operatorID := int64(1)
	r := &AuditLogQuery{
		Page:       1,
		PageSize:   20,
		Action:     "create_user",
		OperatorID: &operatorID,
		StartDate:  "2024-01-01",
		EndDate:    "2024-01-31",
	}
	assert.Equal(t, 1, r.Page)
	assert.Equal(t, 20, r.PageSize)
	assert.Equal(t, "create_user", r.Action)
	assert.Equal(t, &operatorID, r.OperatorID)
	assert.Equal(t, "2024-01-01", r.StartDate)
	assert.Equal(t, "2024-01-31", r.EndDate)
}

func TestCreateMCPServiceRequest_StructFields(t *testing.T) {
	r := &CreateMCPServiceRequest{
		Name:          "test-service",
		DisplayName:   "Test Service",
		Description:   "Test Description",
		EndpointURL:   "https://mcp.test.com",
		TransportType: "sse",
		AuthType:      "bearer",
		AuthConfig:    map[string]string{"token": "test"},
	}
	assert.Equal(t, "test-service", r.Name)
	assert.Equal(t, "Test Service", r.DisplayName)
	assert.Equal(t, "Test Description", r.Description)
	assert.Equal(t, "https://mcp.test.com", r.EndpointURL)
	assert.Equal(t, "sse", r.TransportType)
	assert.Equal(t, "bearer", r.AuthType)
}

func TestUpdateMCPServiceRequest_StructFields(t *testing.T) {
	name := "Updated"
	status := "disabled"
	r := &UpdateMCPServiceRequest{
		DisplayName: &name,
		Status:      &status,
	}
	assert.Equal(t, &name, r.DisplayName)
	assert.Equal(t, &status, r.Status)
}

func TestSetMCPAccessRuleRequest_StructFields(t *testing.T) {
	r := &SetMCPAccessRuleRequest{
		ServiceID:  1,
		TargetType: "user",
		TargetID:   2,
		Allowed:    true,
	}
	assert.Equal(t, int64(1), r.ServiceID)
	assert.Equal(t, "user", r.TargetType)
	assert.Equal(t, int64(2), r.TargetID)
	assert.True(t, r.Allowed)
}
