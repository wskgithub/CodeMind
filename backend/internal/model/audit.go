package model

import (
	"encoding/json"
	"time"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID         int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OperatorID int64           `gorm:"not null;index" json:"operator_id"`
	Action     string          `gorm:"size:50;not null;index" json:"action"`      // 操作类型
	TargetType string          `gorm:"size:50;not null" json:"target_type"`       // 操作目标类型
	TargetID   *int64          `json:"target_id"`
	Detail     json.RawMessage `gorm:"type:jsonb" json:"detail"`                  // 操作详情（变更前后）
	ClientIP   *string         `gorm:"size:45" json:"client_ip"`
	CreatedAt  time.Time       `gorm:"not null;autoCreateTime;index" json:"created_at"`

	// 关联（仅查询时使用）
	Operator *User `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// 审计操作类型常量
const (
	AuditActionCreateUser     = "create_user"
	AuditActionUpdateUser     = "update_user"
	AuditActionDeleteUser     = "delete_user"
	AuditActionDisableUser    = "disable_user"
	AuditActionEnableUser     = "enable_user"
	AuditActionResetPassword  = "reset_password"
	AuditActionImportUsers    = "import_users"
	AuditActionUnlockUser     = "unlock_user"
	AuditActionCreateDept     = "create_department"
	AuditActionUpdateDept     = "update_department"
	AuditActionDeleteDept     = "delete_department"
	AuditActionCreateKey      = "create_api_key"
	AuditActionDeleteKey      = "delete_api_key"
	AuditActionDisableKey     = "disable_api_key"
	AuditActionEnableKey      = "enable_api_key"
	AuditActionCopyKey        = "copy_api_key"
	AuditActionUpdateLimit    = "update_limit"
	AuditActionDeleteLimit    = "delete_limit"
	AuditActionUpdateConfig   = "update_config"
	AuditActionCreateAnnounce = "create_announcement"
	AuditActionUpdateAnnounce = "update_announcement"
	AuditActionDeleteAnnounce = "delete_announcement"
)

// 审计目标类型常量
const (
	AuditTargetUser         = "user"
	AuditTargetDepartment   = "department"
	AuditTargetAPIKey       = "api_key"
	AuditTargetRateLimit    = "rate_limit"
	AuditTargetConfig       = "system_config"
	AuditTargetAnnouncement = "announcement"
)
