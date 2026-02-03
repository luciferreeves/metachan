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

	// Register AniFetch task (weekly) - fetches anime mappings from Fribb list
	err := GlobalTaskManager.RegisterTask(types.Task{
		Name:     "AnimeFetch",
		Interval: 7 * 24 * time.Hour,
		Execute: func() error {
			// Run AniFetch first
			if err := AniFetch(); err != nil {
				return err
			}
			// After AniFetch completes, trigger AniSync in background
			go func() {
				if err := AniSync(); err != nil {
					logger.Log(fmt.Sprintf("AniSync failed: %v", err), logger.LogOptions{
						Level:  logger.Error,
						Prefix: "TaskManager",
					})
					GlobalTaskManager.logTaskExecution("AnimeSync", "error", err.Error())
				} else {
					// Update AnimeSync's LastRun after successful completion
					animeSyncEndTime := time.Now()
					GlobalTaskManager.UpdateTaskLastRun("AnimeSync", animeSyncEndTime)
					GlobalTaskManager.logTaskExecution("AnimeSync", "success", "Task executed successfully")
				}
			}()
			return nil
		},
	})

	if err != nil {
		logger.Log(fmt.Sprintf("Failed to register AnimeFetch task: %v", err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "TaskManager",
		})
	}

	// Register AnimeSync task (triggered automatically after AnimeFetch completes)
	err = GlobalTaskManager.RegisterTask(types.Task{
		Name:        "AnimeSync",
		Interval:    0, // Manual-only - runs after AnimeFetch
		Execute:     AniSync,
		TriggeredBy: "AnimeFetch",
	})

	if err != nil {
		logger.Log(fmt.Sprintf("Failed to register AnimeSync task: %v", err), logger.LogOptions{
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

	// Register GenreSync task (every 7 days)
	err = GlobalTaskManager.RegisterTask(types.Task{
		Name:     "GenreSync",
		Interval: 7 * 24 * time.Hour,
		Execute:  GenreSync,
	})

	if err != nil {
		logger.Log(fmt.Sprintf("Failed to register GenreSync task: %v", err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "TaskManager",
		})
	}
}
