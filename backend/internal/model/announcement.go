package model

import "time"

// Announcement represents a system announcement.
type Announcement struct {
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
	Author    *User     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AuthorID  int64     `gorm:"not null" json:"author_id"`
	Status    int16     `gorm:"not null;default:1" json:"status"`
	Pinned    bool      `gorm:"not null;default:false" json:"pinned"`
}

// TableName returns the table name.
func (Announcement) TableName() string {
	return "announcements"
}
