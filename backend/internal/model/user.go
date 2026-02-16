package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"size:50;not null;uniqueIndex" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`                          // 永远不序列化到 JSON
	DisplayName  string         `gorm:"size:100;not null" json:"display_name"`
	Email        *string        `gorm:"size:255;uniqueIndex" json:"email"`                   // 可为空
	Phone        *string        `gorm:"size:20" json:"phone"`
	AvatarURL    *string        `gorm:"size:500" json:"avatar_url"`
	Role         string         `gorm:"size:20;not null;default:user;index" json:"role"`     // super_admin | dept_manager | user
	DepartmentID *int64         `gorm:"index" json:"department_id"`
	Status       int16          `gorm:"not null;default:1;index" json:"status"`              // 1=启用 0=禁用
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  *string        `gorm:"size:45" json:"last_login_ip"`
	CreatedAt    time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsSuperAdmin 是否超级管理员
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// IsDeptManager 是否部门经理
func (u *User) IsDeptManager() bool {
	return u.Role == RoleDeptManager
}

// IsActive 账号是否启用
func (u *User) IsActive() bool {
	return u.Status == StatusEnabled
}

// 用户角色常量
const (
	RoleSuperAdmin = "super_admin"
	RoleDeptManager = "dept_manager"
	RoleUser       = "user"
)

// 通用状态常量
const (
	StatusDisabled int16 = 0
	StatusEnabled  int16 = 1
)
