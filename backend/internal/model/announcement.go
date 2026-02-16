package model

import "time"

// Announcement 系统公告模型
type Announcement struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"` // 支持 Markdown
	AuthorID  int64     `gorm:"not null" json:"author_id"`
	Status    int16     `gorm:"not null;default:1" json:"status"`  // 1=已发布 0=草稿
	Pinned    bool      `gorm:"not null;default:false" json:"pinned"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`

	// 关联
	Author *User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// TableName 指定表名
func (Announcement) TableName() string {
	return "announcements"
}
