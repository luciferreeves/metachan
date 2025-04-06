package types

import (
	"time"
)

type Task struct {
	Name     string
	Interval time.Duration
	Execute  func() error
	LastRun  time.Time
}

type TaskStatus struct {
	Registered bool
	Running    bool
	LastRun    *time.Time
	NextRun    *time.Time
}
