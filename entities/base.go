package entities

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel extends gorm.Model but hides ID and timestamp fields from JSON
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
