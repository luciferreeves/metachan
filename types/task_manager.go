package types

import (
	"time"
)

type Task struct {
	Name         string
	Interval     time.Duration
	Execute      func() error
	LastRun      time.Time
	Dependencies []string // List of task names that must complete before this task runs
}

type TaskStatus struct {
	Registered bool
	Running    bool
	LastRun    *time.Time
	NextRun    *time.Time
}
