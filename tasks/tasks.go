package tasks

import (
	"fmt"
	"metachan/database"
	"metachan/types"
	"metachan/utils/logger"
	"sync"
	"time"
)

var GlobalTaskManager *TaskManager

func init() {
	GlobalTaskManager = &TaskManager{
		Tasks:    make(map[string]types.Task),
		Tickers:  make(map[string]*time.Ticker),
		Done:     make(map[string]chan bool),
		Mutex:    sync.Mutex{},
		Database: database.DB,
	}

	err := GlobalTaskManager.RegisterTask(types.Task{
		Name:     "AnimeSync",
		Interval: 7 * 24 * time.Hour,
		Execute:  AniSync,
	})

	if err != nil {
		logger.Log(fmt.Sprintf("Failed to register task: %v", err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "TaskManager",
		})
	}
}
