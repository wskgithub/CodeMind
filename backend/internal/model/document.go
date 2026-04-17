package model

import "time"

// Document 表示一篇使用文档.
type Document struct {
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at"`
	Slug        string     `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Title       string     `gorm:"size:200;not null" json:"title"`
	Subtitle    string     `gorm:"size:500" json:"subtitle"`
	Icon        string     `gorm:"size:100" json:"icon"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	SortOrder   int        `gorm:"not null;default:0" json:"sort_order"`
	IsPublished bool       `gorm:"not null;default:false" json:"is_published"`
}

// TableName 返回数据库表名。
func (Document) TableName() string {
	return "documents"
}

// DocumentListItem 文档列表项（不含正文）.
type DocumentListItem struct {
	UpdatedAt   time.Time `json:"updated_at"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Icon        string    `json:"icon"`
	ID          int64     `json:"id"`
	SortOrder   int       `json:"sort_order"`
	IsPublished bool      `json:"is_published"`
}
