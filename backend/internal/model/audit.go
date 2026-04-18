package model

import (
	"encoding/json"
	"time"
)

// AuditLog records user operations for compliance.
type AuditLog struct {
	CreatedAt  time.Time       `gorm:"not null;autoCreateTime;index" json:"created_at"`
	TargetID   *int64          `json:"target_id"`
	ClientIP   *string         `gorm:"size:45" json:"client_ip"`
	Operator   *User           `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
	Action     string          `gorm:"size:50;not null;index" json:"action"`
	TargetType string          `gorm:"size:50;not null" json:"target_type"`
	Detail     json.RawMessage `gorm:"type:jsonb" json:"detail"`
	ID         int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OperatorID int64           `gorm:"not null;index" json:"operator_id"`
}

// TableName returns the database table name.
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Audit log action type constants.
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

// Audit log target type constants.
const (
	AuditTargetUser         = "user"
	AuditTargetDepartment   = "department"
	AuditTargetAPIKey       = "api_key"
	AuditTargetRateLimit    = "rate_limit"
	AuditTargetConfig       = "system_config"
	AuditTargetAnnouncement = "announcement"
)
