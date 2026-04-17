package model

import "time"

// Department represents an organizational unit
type Department struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description *string    `gorm:"type:text" json:"description"`
	ManagerID   *int64     `json:"manager_id"`
	ParentID    *int64     `json:"parent_id"`
	Status      int16      `gorm:"not null;default:1" json:"status"`
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`

	Manager  *User         `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Parent   *Department   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Department  `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (Department) TableName() string {
	return "departments"
}
