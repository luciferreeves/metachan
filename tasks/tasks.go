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

	// Register AniSync task (weekly)
	err := GlobalTaskManager.RegisterTask(types.Task{
		Name:     "AnimeSync",
		Interval: 7 * 24 * time.Hour,
		Execute:  AniSync,
	})

	if err != nil {
		logger.Log(fmt.Sprintf("Failed to register AniSync task: %v", err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "TaskManager",
		})
	}

	// Register AnimeUpdate task (every 15 minutes)
	err = GlobalTaskManager.RegisterTask(types.Task{
		Name:     "AnimeUpdate",
		Interval: 15 * time.Minute,
		Execute:  AnimeUpdate,
	})

	if err != nil {
		logger.Log(fmt.Sprintf("Failed to register AnimeUpdate task: %v", err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "TaskManager",
		})
	}
}
