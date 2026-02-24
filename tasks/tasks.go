package tasks

import (
	"metachan/config"
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
		Name:     "ProducerSync",
		Interval: 7 * 24 * time.Hour,
		Execute:  ProducerSync,
		OnResume: ResumeProducerEnrichment,
	})

	if err != nil {
		logger.Errorf("TaskManager", "Failed to register ProducerSync task: %v", err)
	}

	err = GlobalTaskManager.RegisterTask(types.Task{
		Name:     "GenreSync",
		Interval: 7 * 24 * time.Hour,
		Execute:  GenreSync,
	})

	if err != nil {
		logger.Errorf("TaskManager", "Failed to register GenreSync task: %v", err)
	}

	err = GlobalTaskManager.RegisterTask(types.Task{
		Name:         "AnimeFetch",
		Interval:     7 * 24 * time.Hour,
		Dependencies: []string{"ProducerSync", "GenreSync"},
		Execute:      AniFetch,
	})

	if err != nil {
		logger.Errorf("TaskManager", "Failed to register AnimeFetch task: %v", err)
	}

	if config.Sync.AniSync {
		err = GlobalTaskManager.RegisterTask(types.Task{
			Name:         "AnimeSync",
			Interval:     0,
			Execute:      AniSync,
			OnResume:     ResumeAnimeSync,
			Dependencies: []string{"AnimeFetch"},
		})

		if err != nil {
			logger.Errorf("TaskManager", "Failed to register AnimeSync task: %v", err)
		}

		err = GlobalTaskManager.RegisterTask(types.Task{
			Name:         "CharacterSync",
			Interval:     0,
			Execute:      CharacterSync,
			OnResume:     ResumeCharacterEnrichment,
			Dependencies: []string{"AnimeSync"},
		})
		if err != nil {
			logger.Errorf("TaskManager", "Failed to register CharacterSync task: %v", err)
		}

		err = GlobalTaskManager.RegisterTask(types.Task{
			Name:         "PersonSync",
			Interval:     0,
			Execute:      PersonSync,
			OnResume:     ResumePersonEnrichment,
			Dependencies: []string{"CharacterSync"},
		})
		if err != nil {
			logger.Errorf("TaskManager", "Failed to register PersonSync task: %v", err)
		}
	}

	err = GlobalTaskManager.RegisterTask(types.Task{
		Name:     "AnimeUpdate",
		Interval: 15 * time.Minute,
		Execute:  AnimeUpdate,
	})

	if err != nil {
		logger.Errorf("TaskManager", "Failed to register AnimeUpdate task: %v", err)
	}
}
