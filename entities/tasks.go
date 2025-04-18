package entities

import (
	"time"

	"gorm.io/gorm"
)

type TaskLog struct {
	gorm.Model
	TaskName   string `gorm:"index"`
	Status     string // 'success', 'error', 'running'
	Message    string // error message if any
	ExecutedAt time.Time
}
