package tasks

import (
	"metachan/types"
	"metachan/utils/logger"
	"sync"
	"time"
)

var GlobalTaskManager *TaskManager

func init() {
	GlobalTaskManager = &TaskManager{
		Tasks:   make(map[string]types.Task),
		Tickers: make(map[string]*time.Ticker),
		Done:    make(map[string]chan bool),
		Mutex:   sync.Mutex{},
	}

	err := GlobalTaskManager.RegisterTask(types.Task{
		Name:     "AnimeFetch",
		Interval: 7 * 24 * time.Hour,
		Execute:  AniFetch,
	})

	if err != nil {
		logger.Errorf("TaskManager", "Failed to register AnimeFetch task: %v", err)
	}
}
