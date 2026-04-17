package model

import "time"

// APIKey represents an API key for LLM proxy access.
type APIKey struct {
	CreatedAt    time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	User         *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	KeyPrefix    string     `gorm:"size:20;not null;index" json:"key_prefix"`
	KeyHash      string     `gorm:"size:255;not null;uniqueIndex" json:"-"`
	KeyEncrypted string     `gorm:"size:255" json:"-"`
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64      `gorm:"not null;index" json:"user_id"`
	Status       int16      `gorm:"not null;default:1" json:"status"`
}

// TableName 返回数据库表名。
func (APIKey) TableName() string {
	return "api_keys"
}

// IsActive 判断 API Key 是否处于激活状态。
func (k *APIKey) IsActive() bool {
	return k.Status == StatusEnabled
}

// IsExpired 判断 API Key 是否已过期。
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}
