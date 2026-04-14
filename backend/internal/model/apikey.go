package model

import "time"

// APIKey API 密钥模型
type APIKey struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"not null;index" json:"user_id"`
	Name       string     `gorm:"size:100;not null" json:"name"`
	KeyPrefix  string     `gorm:"size:20;not null;index" json:"key_prefix"`   // 前缀用于展示（如 cm-a1b2c3d4）
	KeyHash     string     `gorm:"size:255;not null;uniqueIndex" json:"-"`     // SHA-256 哈希，不暴露
	KeyEncrypted string     `gorm:"size:255" json:"-"`                          // AES-256-GCM 加密的完整 Key，用于复制功能
	Status     int16      `gorm:"not null;default:1" json:"status"`           // 1=启用 0=禁用
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// IsActive 密钥是否可用
func (k *APIKey) IsActive() bool {
	return k.Status == StatusEnabled
}

// IsExpired 密钥是否已过期
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false // 永不过期
	}
	return time.Now().After(*k.ExpiresAt)
}
