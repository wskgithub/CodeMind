package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a platform user.
type User struct {
	CreatedAt       time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DepartmentID    *int64         `gorm:"index" json:"department_id"`
	LockedUntil     *time.Time     `json:"locked_until"`
	Email           *string        `gorm:"size:255;uniqueIndex" json:"email"`
	Phone           *string        `gorm:"size:20" json:"phone"`
	AvatarURL       *string        `gorm:"size:500" json:"avatar_url"`
	Department      *Department    `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	LastLoginFailAt *time.Time     `json:"last_login_fail_at"`
	LastLoginAt     *time.Time     `json:"last_login_at"`
	LastLoginIP     *string        `gorm:"size:45" json:"last_login_ip"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Role            string         `gorm:"size:20;not null;default:user;index" json:"role"`
	PasswordHash    string         `gorm:"size:255;not null" json:"-"`
	Username        string         `gorm:"size:50;not null;uniqueIndex" json:"username"`
	DisplayName     string         `gorm:"size:100;not null" json:"display_name"`
	LoginFailCount  int            `gorm:"not null;default:0" json:"login_fail_count"`
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Status          int16          `gorm:"not null;default:1;index" json:"status"`
}

// TableName 返回数据库表名。
func (User) TableName() string {
	return "users"
}

// IsSuperAdmin 判断用户是否为超级管理员。
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// IsDeptManager 判断用户是否为部门管理员。
func (u *User) IsDeptManager() bool {
	return u.Role == RoleDeptManager
}

// IsActive 判断用户是否处于激活状态。
func (u *User) IsActive() bool {
	return u.Status == StatusEnabled
}

// IsLocked 判断用户是否被锁定。
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// GetRemainingLockTime returns remaining lock time in seconds.
func (u *User) GetRemainingLockTime() int64 {
	if u.LockedUntil == nil {
		return 0
	}
	remaining := u.LockedUntil.Unix() - time.Now().Unix()
	if remaining < 0 {
		return 0
	}
	return remaining
}

// 用户角色常量。
const (
	RoleSuperAdmin  = "super_admin"
	RoleDeptManager = "dept_manager"
	RoleUser        = "user"
)

// 用户状态常量。
const (
	StatusDisabled int16 = 0
	StatusEnabled  int16 = 1
)
