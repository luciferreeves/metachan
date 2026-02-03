package types

import (
	"time"
)

type Task struct {
	Name        string
	Interval    time.Duration
	Execute     func() error
	LastRun     time.Time
	TriggeredBy string // Name of parent task that triggers this task (for manual tasks)
}

type TaskStatus struct {
	Registered bool
	Running    bool
	LastRun    *time.Time
	NextRun    *time.Time
}
