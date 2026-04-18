package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ==================== User model tests ====================

func TestUser_IsSuperAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"super admin", RoleSuperAdmin, true},
		{"department manager", RoleDeptManager, false},
		{"regular user", RoleUser, false},
		{"empty role", "", false},
		{"other role", "admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			assert.Equal(t, tt.expected, u.IsSuperAdmin())
		})
	}
}

func TestUser_IsDeptManager(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"super admin", RoleSuperAdmin, false},
		{"department manager", RoleDeptManager, true},
		{"regular user", RoleUser, false},
		{"empty role", "", false},
		{"other role", "manager", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			assert.Equal(t, tt.expected, u.IsDeptManager())
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   int16
		expected bool
	}{
		{"enabled status", StatusEnabled, true},
		{"disabled status", StatusDisabled, false},
		{"other status 2", int16(2), false},
		{"other status -1", int16(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Status: tt.status}
			assert.Equal(t, tt.expected, u.IsActive())
		})
	}
}

func TestUser_IsLocked(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name      string
		lockUntil *time.Time
		expected  bool
	}{
		{"not locked - nil", nil, false},
		{"locked - future time", &future, true},
		{"lock expired - past time", &past, false},
		{"lock just expired - now", &now, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{LockedUntil: tt.lockUntil}
			assert.Equal(t, tt.expected, u.IsLocked())
		})
	}
}

func TestUser_GetRemainingLockTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		lockUntil   *time.Time
		expectedMin int64
		expectedMax int64
	}{
		{"not locked - nil", nil, 0, 0},
		{"lock expired", func() *time.Time { t := now.Add(-time.Hour); return &t }(), 0, 0},
		{"locked for 30 minutes", func() *time.Time { t := now.Add(30 * time.Minute); return &t }(), 29 * 60, 31 * 60},
		{"locked for 1 hour", func() *time.Time { t := now.Add(time.Hour); return &t }(), 59 * 60, 61 * 60},
		{"locked for 1 second", func() *time.Time { t := now.Add(time.Second); return &t }(), 0, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{LockedUntil: tt.lockUntil}
			result := u.GetRemainingLockTime()
			assert.GreaterOrEqual(t, result, tt.expectedMin)
			assert.LessOrEqual(t, result, tt.expectedMax)
		})
	}
}

// ==================== APIKey model tests ====================

func TestAPIKey_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   int16
		expected bool
	}{
		{"enabled status", StatusEnabled, true},
		{"disabled status", StatusDisabled, false},
		{"other status", int16(2), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &APIKey{Status: tt.status}
			assert.Equal(t, tt.expected, k.IsActive())
		})
	}
}

func TestAPIKey_IsExpired(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  bool
	}{
		{"never expires - nil", nil, false},
		{"expired - past time", &past, true},
		{"not expired - future time", &future, false},
		{"just expired - now", &now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &APIKey{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.expected, k.IsExpired())
		})
	}
}

// ==================== RateLimit model tests ====================

func TestPeriodLabel(t *testing.T) {
	tests := []struct {
		hours    int
		expected string
	}{
		{24, PeriodDaily},
		{168, PeriodWeekly},
		{720, PeriodMonthly},
		{0, PeriodCustom},
		{1, PeriodCustom},
		{23, PeriodCustom},
		{25, PeriodCustom},
		{100, PeriodCustom},
		{-1, PeriodCustom},
		{-24, PeriodCustom},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := PeriodLabel(tt.hours)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPeriodHoursFromLabel(t *testing.T) {
	tests := []struct {
		label    string
		expected int
	}{
		{PeriodDaily, 24},
		{PeriodWeekly, 168},
		{PeriodMonthly, 720},
		{PeriodCustom, 24},
		{"", 24},
		{"invalid", 24},
		{"yearly", 24},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := PeriodHoursFromLabel(tt.label)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRateLimit_EffectiveHours(t *testing.T) {
	tests := []struct {
		name        string
		periodHours int
		period      string
		expected    int
	}{
		{"PeriodHours set to 24", 24, "", 24},
		{"PeriodHours set to 168", 168, "", 168},
		{"PeriodHours set to 720", 720, "", 720},
		{"PeriodHours set to 48", 48, "", 48},
		{"PeriodHours set to negative", -1, "", 24},
		{"PeriodHours is 0 - use daily", 0, PeriodDaily, 24},
		{"PeriodHours is 0 - use weekly", 0, PeriodWeekly, 168},
		{"PeriodHours is 0 - use monthly", 0, PeriodMonthly, 720},
		{"PeriodHours is 0 - use custom", 0, PeriodCustom, 24},
		{"PeriodHours is 0 - empty string", 0, "", 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RateLimit{
				PeriodHours: tt.periodHours,
				Period:      tt.period,
			}
			assert.Equal(t, tt.expected, r.EffectiveHours())
		})
	}
}

// ==================== Constant validation tests ====================

func TestUserRoleConstants(t *testing.T) {
	assert.Equal(t, "super_admin", RoleSuperAdmin)
	assert.Equal(t, "dept_manager", RoleDeptManager)
	assert.Equal(t, "user", RoleUser)
}

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, int16(0), StatusDisabled)
	assert.Equal(t, int16(1), StatusEnabled)
}

func TestRateLimitConstants(t *testing.T) {
	// Target type constants
	assert.Equal(t, "global", TargetTypeGlobal)
	assert.Equal(t, "department", TargetTypeDepartment)
	assert.Equal(t, "user", TargetTypeUser)

	// Period constants
	assert.Equal(t, "daily", PeriodDaily)
	assert.Equal(t, "weekly", PeriodWeekly)
	assert.Equal(t, "monthly", PeriodMonthly)
	assert.Equal(t, "custom", PeriodCustom)
}

func TestLLMBackendConstants(t *testing.T) {
	assert.Equal(t, 0, LLMBackendDisabled)
	assert.Equal(t, 1, LLMBackendEnabled)
	assert.Equal(t, 2, LLMBackendDraining)
}

func TestMCPConstants(t *testing.T) {
	// Service status
	assert.Equal(t, "enabled", MCPServiceEnabled)
	assert.Equal(t, "disabled", MCPServiceDisabled)

	// Transport type
	assert.Equal(t, "sse", MCPTransportSSE)
	assert.Equal(t, "streamable-http", MCPTransportStreamableHTTP)

	// Auth type
	assert.Equal(t, "none", MCPAuthNone)
	assert.Equal(t, "bearer", MCPAuthBearer)
	assert.Equal(t, "header", MCPAuthHeader)

	// Access rule target types
	assert.Equal(t, "user", MCPTargetUser)
	assert.Equal(t, "department", MCPTargetDepartment)
	assert.Equal(t, "role", MCPTargetRole)
}

func TestSystemConfigConstants(t *testing.T) {
	assert.Equal(t, "llm.base_url", ConfigLLMBaseURL)
	assert.Equal(t, "llm.api_key", ConfigLLMAPIKey)
	assert.Equal(t, "llm.models", ConfigLLMModels)
	assert.Equal(t, "llm.default_model", ConfigLLMDefaultModel)
	assert.Equal(t, "system.max_keys_per_user", ConfigMaxKeysPerUser)
	assert.Equal(t, "system.default_concurrency", ConfigDefaultConcurrency)
	assert.Equal(t, "system.force_change_password", ConfigForceChangePwd)
}

// ==================== Model struct validation tests ====================

func TestUser_StructFields(t *testing.T) {
	email := "test@example.com"
	phone := "1234567890"
	avatarURL := "https://example.com/avatar.png"
	departmentID := int64(1)
	lastLoginAt := time.Now()
	lastLoginIP := "127.0.0.1"
	lockedUntil := time.Now().Add(time.Hour)
	lastLoginFailAt := time.Now()

	u := &User{
		ID:              1,
		Username:        "testuser",
		PasswordHash:    "hashedpassword",
		DisplayName:     "Test User",
		Email:           &email,
		Phone:           &phone,
		AvatarURL:       &avatarURL,
		Role:            RoleUser,
		DepartmentID:    &departmentID,
		Status:          StatusEnabled,
		LastLoginAt:     &lastLoginAt,
		LastLoginIP:     &lastLoginIP,
		LoginFailCount:  0,
		LockedUntil:     &lockedUntil,
		LastLoginFailAt: &lastLoginFailAt,
	}

	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, "testuser", u.Username)
	assert.Equal(t, "hashedpassword", u.PasswordHash)
	assert.Equal(t, "Test User", u.DisplayName)
	assert.Equal(t, &email, u.Email)
	assert.Equal(t, &phone, u.Phone)
	assert.Equal(t, &avatarURL, u.AvatarURL)
	assert.Equal(t, RoleUser, u.Role)
	assert.Equal(t, &departmentID, u.DepartmentID)
	assert.Equal(t, StatusEnabled, u.Status)
}

func TestAPIKey_StructFields(t *testing.T) {
	lastUsedAt := time.Now()
	expiresAt := time.Now().Add(time.Hour * 24)

	key := &APIKey{
		ID:         1,
		UserID:     1,
		Name:       "Test Key",
		KeyPrefix:  "cm-a1b2c3",
		KeyHash:    "hash123",
		Status:     StatusEnabled,
		LastUsedAt: &lastUsedAt,
		ExpiresAt:  &expiresAt,
	}

	assert.Equal(t, int64(1), key.ID)
	assert.Equal(t, int64(1), key.UserID)
	assert.Equal(t, "Test Key", key.Name)
	assert.Equal(t, "cm-a1b2c3", key.KeyPrefix)
	assert.Equal(t, "hash123", key.KeyHash)
	assert.Equal(t, StatusEnabled, key.Status)
	assert.Equal(t, &lastUsedAt, key.LastUsedAt)
	assert.Equal(t, &expiresAt, key.ExpiresAt)
}

func TestDepartment_StructFields(t *testing.T) {
	description := "Test Department"
	managerID := int64(1)
	parentID := int64(0)

	dept := &Department{
		ID:          1,
		Name:        "Test Dept",
		Description: &description,
		ManagerID:   &managerID,
		ParentID:    &parentID,
		Status:      StatusEnabled,
	}

	assert.Equal(t, int64(1), dept.ID)
	assert.Equal(t, "Test Dept", dept.Name)
	assert.Equal(t, &description, dept.Description)
	assert.Equal(t, &managerID, dept.ManagerID)
	assert.Equal(t, &parentID, dept.ParentID)
	assert.Equal(t, StatusEnabled, dept.Status)
}

func TestRateLimit_StructFields(t *testing.T) {
	rl := &RateLimit{
		ID:             1,
		TargetType:     TargetTypeUser,
		TargetID:       1,
		Period:         PeriodDaily,
		PeriodHours:    24,
		MaxTokens:      100000,
		MaxRequests:    1000,
		MaxConcurrency: 5,
		AlertThreshold: 80,
		Status:         StatusEnabled,
	}

	assert.Equal(t, int64(1), rl.ID)
	assert.Equal(t, TargetTypeUser, rl.TargetType)
	assert.Equal(t, int64(1), rl.TargetID)
	assert.Equal(t, PeriodDaily, rl.Period)
	assert.Equal(t, 24, rl.PeriodHours)
	assert.Equal(t, int64(100000), rl.MaxTokens)
	assert.Equal(t, 1000, rl.MaxRequests)
	assert.Equal(t, 5, rl.MaxConcurrency)
	assert.Equal(t, int16(80), rl.AlertThreshold)
	assert.Equal(t, StatusEnabled, rl.Status)
}

func TestAnnouncement_StructFields(t *testing.T) {
	ann := &Announcement{
		ID:       1,
		Title:    "Test Announcement",
		Content:  "This is a test announcement",
		AuthorID: 1,
		Status:   StatusEnabled,
		Pinned:   true,
	}

	assert.Equal(t, int64(1), ann.ID)
	assert.Equal(t, "Test Announcement", ann.Title)
	assert.Equal(t, "This is a test announcement", ann.Content)
	assert.Equal(t, int64(1), ann.AuthorID)
	assert.Equal(t, StatusEnabled, ann.Status)
	assert.True(t, ann.Pinned)
}

func TestLLMBackend_StructFields(t *testing.T) {
	backend := &LLMBackend{
		ID:                   1,
		Name:                 "test-backend",
		DisplayName:          "Test Backend",
		BaseURL:              "https://api.test.com",
		APIKey:               "sk-test123",
		Format:               "openai",
		Weight:               100,
		MaxConcurrency:       100,
		Status:               LLMBackendEnabled,
		HealthCheckURL:       "https://api.test.com/health",
		TimeoutSeconds:       300,
		StreamTimeoutSeconds: 600,
		ModelPatterns:        "gpt-*,claude-*",
	}

	assert.Equal(t, int64(1), backend.ID)
	assert.Equal(t, "test-backend", backend.Name)
	assert.Equal(t, "Test Backend", backend.DisplayName)
	assert.Equal(t, "https://api.test.com", backend.BaseURL)
	assert.Equal(t, "sk-test123", backend.APIKey)
	assert.Equal(t, "openai", backend.Format)
	assert.Equal(t, 100, backend.Weight)
	assert.Equal(t, 100, backend.MaxConcurrency)
	assert.Equal(t, int16(LLMBackendEnabled), backend.Status)
}

func TestMCPService_StructFields(t *testing.T) {
	service := &MCPService{
		ID:            1,
		Name:          "test-service",
		DisplayName:   "Test Service",
		Description:   "Test Description",
		EndpointURL:   "https://mcp.test.com",
		TransportType: MCPTransportSSE,
		Status:        MCPServiceEnabled,
		AuthType:      MCPAuthBearer,
	}

	assert.Equal(t, int64(1), service.ID)
	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "Test Service", service.DisplayName)
	assert.Equal(t, "Test Description", service.Description)
	assert.Equal(t, "https://mcp.test.com", service.EndpointURL)
	assert.Equal(t, MCPTransportSSE, service.TransportType)
	assert.Equal(t, MCPServiceEnabled, service.Status)
	assert.Equal(t, MCPAuthBearer, service.AuthType)
}

func TestMCPAccessRule_StructFields(t *testing.T) {
	rule := &MCPAccessRule{
		ID:         1,
		ServiceID:  1,
		TargetType: MCPTargetUser,
		TargetID:   1,
		Allowed:    true,
	}

	assert.Equal(t, int64(1), rule.ID)
	assert.Equal(t, int64(1), rule.ServiceID)
	assert.Equal(t, MCPTargetUser, rule.TargetType)
	assert.Equal(t, int64(1), rule.TargetID)
	assert.True(t, rule.Allowed)
}

func TestSystemConfig_StructFields(t *testing.T) {
	description := "Test Config"
	config := &SystemConfig{
		ID:          1,
		ConfigKey:   "test.key",
		ConfigValue: `{"value": "test"}`,
		Description: &description,
	}

	assert.Equal(t, int64(1), config.ID)
	assert.Equal(t, "test.key", config.ConfigKey)
	assert.Equal(t, `{"value": "test"}`, config.ConfigValue)
	assert.Equal(t, &description, config.Description)
}

func TestTokenUsage_StructFields(t *testing.T) {
	durationMs := 1000
	usage := &TokenUsage{
		ID:               1,
		UserID:           1,
		APIKeyID:         1,
		Model:            "gpt-4",
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		RequestType:      "chat_completion",
		DurationMs:       &durationMs,
	}

	assert.Equal(t, int64(1), usage.ID)
	assert.Equal(t, int64(1), usage.UserID)
	assert.Equal(t, int64(1), usage.APIKeyID)
	assert.Equal(t, "gpt-4", usage.Model)
	assert.Equal(t, 100, usage.PromptTokens)
	assert.Equal(t, 50, usage.CompletionTokens)
	assert.Equal(t, 150, usage.TotalTokens)
	assert.Equal(t, "chat_completion", usage.RequestType)
	assert.Equal(t, &durationMs, usage.DurationMs)
}

func TestTokenUsageDaily_StructFields(t *testing.T) {
	today := time.Now()
	usage := &TokenUsageDaily{
		ID:               1,
		UserID:           1,
		UsageDate:        today,
		PromptTokens:     1000,
		CompletionTokens: 500,
		TotalTokens:      1500,
		RequestCount:     10,
	}

	assert.Equal(t, int64(1), usage.ID)
	assert.Equal(t, int64(1), usage.UserID)
	assert.Equal(t, today, usage.UsageDate)
	assert.Equal(t, int64(1000), usage.PromptTokens)
	assert.Equal(t, int64(500), usage.CompletionTokens)
	assert.Equal(t, int64(1500), usage.TotalTokens)
	assert.Equal(t, 10, usage.RequestCount)
}

func TestRequestLog_StructFields(t *testing.T) {
	model := "gpt-4"
	errorMsg := "test error"
	clientIP := "127.0.0.1"
	userAgent := "test-agent"
	durationMs := 1000

	log := &RequestLog{
		ID:           1,
		UserID:       1,
		APIKeyID:     1,
		RequestType:  "chat_completion",
		Model:        &model,
		StatusCode:   200,
		ErrorMessage: &errorMsg,
		ClientIP:     &clientIP,
		UserAgent:    &userAgent,
		DurationMs:   &durationMs,
	}

	assert.Equal(t, int64(1), log.ID)
	assert.Equal(t, int64(1), log.UserID)
	assert.Equal(t, int64(1), log.APIKeyID)
	assert.Equal(t, "chat_completion", log.RequestType)
	assert.Equal(t, &model, log.Model)
	assert.Equal(t, 200, log.StatusCode)
	assert.Equal(t, &errorMsg, log.ErrorMessage)
	assert.Equal(t, &clientIP, log.ClientIP)
	assert.Equal(t, &userAgent, log.UserAgent)
	assert.Equal(t, &durationMs, log.DurationMs)
}

// ==================== Audit model tests ====================

func TestAuditConstants(t *testing.T) {
	// Audit action type constants
	assert.Equal(t, "create_user", AuditActionCreateUser)
	assert.Equal(t, "update_user", AuditActionUpdateUser)
	assert.Equal(t, "delete_user", AuditActionDeleteUser)
	assert.Equal(t, "disable_user", AuditActionDisableUser)
	assert.Equal(t, "enable_user", AuditActionEnableUser)
	assert.Equal(t, "reset_password", AuditActionResetPassword)
	assert.Equal(t, "import_users", AuditActionImportUsers)
	assert.Equal(t, "unlock_user", AuditActionUnlockUser)
	assert.Equal(t, "create_department", AuditActionCreateDept)
	assert.Equal(t, "update_department", AuditActionUpdateDept)
	assert.Equal(t, "delete_department", AuditActionDeleteDept)
	assert.Equal(t, "create_api_key", AuditActionCreateKey)
	assert.Equal(t, "delete_api_key", AuditActionDeleteKey)
	assert.Equal(t, "disable_api_key", AuditActionDisableKey)
	assert.Equal(t, "enable_api_key", AuditActionEnableKey)
	assert.Equal(t, "update_limit", AuditActionUpdateLimit)
	assert.Equal(t, "delete_limit", AuditActionDeleteLimit)
	assert.Equal(t, "update_config", AuditActionUpdateConfig)
	assert.Equal(t, "create_announcement", AuditActionCreateAnnounce)
	assert.Equal(t, "update_announcement", AuditActionUpdateAnnounce)
	assert.Equal(t, "delete_announcement", AuditActionDeleteAnnounce)

	// Audit target type constants
	assert.Equal(t, "user", AuditTargetUser)
	assert.Equal(t, "department", AuditTargetDepartment)
	assert.Equal(t, "api_key", AuditTargetAPIKey)
	assert.Equal(t, "rate_limit", AuditTargetRateLimit)
	assert.Equal(t, "system_config", AuditTargetConfig)
	assert.Equal(t, "announcement", AuditTargetAnnouncement)
}

func TestAuditLog_StructFields(t *testing.T) {
	targetID := int64(1)
	clientIP := "127.0.0.1"
	detail := json.RawMessage(`{"before": "test", "after": "updated"}`)

	log := &AuditLog{
		ID:         1,
		OperatorID: 2,
		Action:     AuditActionCreateUser,
		TargetType: AuditTargetUser,
		TargetID:   &targetID,
		Detail:     detail,
		ClientIP:   &clientIP,
	}

	assert.Equal(t, int64(1), log.ID)
	assert.Equal(t, int64(2), log.OperatorID)
	assert.Equal(t, AuditActionCreateUser, log.Action)
	assert.Equal(t, AuditTargetUser, log.TargetType)
	assert.Equal(t, &targetID, log.TargetID)
	assert.Equal(t, detail, log.Detail)
	assert.Equal(t, &clientIP, log.ClientIP)
}

// ==================== MCP auth config struct tests ====================

func TestMCPAuthConfigBearer_StructFields(t *testing.T) {
	config := MCPAuthConfigBearer{
		Token: "test-token-123",
	}
	assert.Equal(t, "test-token-123", config.Token)
}

func TestMCPAuthConfigHeader_StructFields(t *testing.T) {
	config := MCPAuthConfigHeader{
		HeaderName:  "X-Custom-Auth",
		HeaderValue: "secret-value",
	}
	assert.Equal(t, "X-Custom-Auth", config.HeaderName)
	assert.Equal(t, "secret-value", config.HeaderValue)
}

// ==================== Edge case tests ====================

func TestUser_IsLocked_Boundary(t *testing.T) {
	now := time.Now()

	// Edge case: exactly 1 second in the future
	oneSecondFuture := now.Add(time.Second)
	u := &User{LockedUntil: &oneSecondFuture}
	assert.True(t, u.IsLocked())

	// Edge case: exactly 1 second in the past
	oneSecondPast := now.Add(-time.Second)
	u2 := &User{LockedUntil: &oneSecondPast}
	assert.False(t, u2.IsLocked())
}

func TestAPIKey_IsExpired_Boundary(t *testing.T) {
	now := time.Now()

	// Edge case: exactly 1 second in the future
	oneSecondFuture := now.Add(time.Second)
	k := &APIKey{ExpiresAt: &oneSecondFuture}
	assert.False(t, k.IsExpired())

	// Edge case: exactly 1 second in the past
	oneSecondPast := now.Add(-time.Second)
	k2 := &APIKey{ExpiresAt: &oneSecondPast}
	assert.True(t, k2.IsExpired())
}

func TestRateLimit_PeriodLabel_AllPossibleHours(t *testing.T) {
	// Test all possible hour values
	testCases := map[int]string{
		-720: PeriodCustom,
		-168: PeriodCustom,
		-24:  PeriodCustom,
		-1:   PeriodCustom,
		0:    PeriodCustom,
		1:    PeriodCustom,
		12:   PeriodCustom,
		23:   PeriodCustom,
		24:   PeriodDaily,
		25:   PeriodCustom,
		48:   PeriodCustom,
		72:   PeriodCustom,
		100:  PeriodCustom,
		167:  PeriodCustom,
		168:  PeriodWeekly,
		169:  PeriodCustom,
		500:  PeriodCustom,
		719:  PeriodCustom,
		720:  PeriodMonthly,
		721:  PeriodCustom,
		1000: PeriodCustom,
		8760: PeriodCustom, // one year
	}

	for hours, expected := range testCases {
		result := PeriodLabel(hours)
		assert.Equal(t, expected, result, "hours=%d", hours)
	}
}

func TestRateLimit_EffectiveHours_Boundary(t *testing.T) {
	// Edge case: PeriodHours is 1 (minimum positive value)
	rl1 := &RateLimit{PeriodHours: 1, Period: PeriodDaily}
	assert.Equal(t, 1, rl1.EffectiveHours())

	// Edge case: PeriodHours is negative (falls back to Period label derivation)
	rl2 := &RateLimit{PeriodHours: -1, Period: PeriodDaily}
	assert.Equal(t, 24, rl2.EffectiveHours())

	// Edge case: PeriodHours is 0, inferred from Period
	rl3 := &RateLimit{PeriodHours: 0, Period: PeriodWeekly}
	assert.Equal(t, 168, rl3.EffectiveHours())

	// Edge case: PeriodHours is very large
	rl4 := &RateLimit{PeriodHours: 9999, Period: PeriodDaily}
	assert.Equal(t, 9999, rl4.EffectiveHours())
}

func TestUser_GetRemainingLockTime_Boundary(t *testing.T) {
	now := time.Now()

	// Edge case: 1 second remaining
	oneSecondFuture := now.Add(time.Second)
	u := &User{LockedUntil: &oneSecondFuture}
	remaining := u.GetRemainingLockTime()
	assert.GreaterOrEqual(t, remaining, int64(0))
	assert.LessOrEqual(t, remaining, int64(2))

	// Edge case: 0 seconds remaining (just expired)
	u2 := &User{LockedUntil: &now}
	remaining2 := u2.GetRemainingLockTime()
	assert.Equal(t, int64(0), remaining2)

	// Edge case: very long time remaining
	longFuture := now.Add(time.Hour * 24 * 365) // 1 year
	u3 := &User{LockedUntil: &longFuture}
	remaining3 := u3.GetRemainingLockTime()
	expectedSeconds := int64(24 * 365 * 3600)
	assert.GreaterOrEqual(t, remaining3, expectedSeconds-1)
	assert.LessOrEqual(t, remaining3, expectedSeconds+1)
}
