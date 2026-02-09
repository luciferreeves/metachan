package entities

import (
	"time"
)

type TaskLog struct {
	BaseModel
	TaskName   string    `gorm:"index" json:"task_name,omitempty"`
	Status     string    `json:"status,omitempty"`
	Message    string    `json:"message,omitempty"`
	ExecutedAt time.Time `json:"executed_at,omitempty"`
}

type TaskStatus struct {
	BaseModel
	TaskName    string    `gorm:"uniqueIndex;not null" json:"task_name"`
	IsCompleted bool      `gorm:"default:false" json:"is_completed,omitempty"`
	LastRunAt   time.Time `json:"last_run_at,omitempty"`
}
