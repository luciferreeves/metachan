package types

import "time"

type Task struct {
	Name         string
	Interval     time.Duration
	Execute      func() error
	OnResume     func()
	Dependencies []string
}

type TaskStatus struct {
	Registered bool
	Running    bool
	LastRun    *time.Time
	NextRun    *time.Time
}
