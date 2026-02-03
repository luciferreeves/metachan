package tasks

import (
	"fmt"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
	"sync"
	"time"

	"gorm.io/gorm"
)

type TaskManager struct {
	Tasks    map[string]types.Task
	Tickers  map[string]*time.Ticker
	Done     map[string]chan bool
	Mutex    sync.Mutex
	Database *gorm.DB
}

func (tm *TaskManager) RegisterTask(task types.Task) error {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	if _, exists := tm.Tasks[task.Name]; exists {
		return fmt.Errorf("task %s already registered", task.Name)
	}

	tm.Tasks[task.Name] = task
	logger.Log(fmt.Sprintf("Task %s registered", task.Name), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "TaskManager",
	})

	return nil
}

func (tm *TaskManager) shouldExecuteTask(taskName string, interval time.Duration) (bool, error) {
	var lastLog entities.TaskLog

	if err := tm.Database.Where("task_name = ?", taskName).Order("executed_at desc").First(&lastLog).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return true, nil
		}

		return false, err
	}

	elapsed := time.Since(lastLog.ExecutedAt)
	return elapsed >= interval, nil
}

func (tm *TaskManager) logTaskExecution(taskName, status, message string) {
	logEntry := entities.TaskLog{
		TaskName:   taskName,
		Status:     status,
		Message:    message,
		ExecutedAt: time.Now(),
	}

	if err := tm.Database.Create(&logEntry).Error; err != nil {
		logger.Log(fmt.Sprintf("Failed to log task execution for %s: %v", taskName, err), logger.LogOptions{
			Level:  logger.Warn,
			Prefix: "TaskManager",
		})
	}
}

func (tm *TaskManager) StartTask(taskName string) {
	tm.Mutex.Lock()
	task, exists := tm.Tasks[taskName]
	tm.Mutex.Unlock()
	if !exists {
		logger.Log(fmt.Sprintf("Task %s not found", taskName), logger.LogOptions{
			Level:  logger.Warn,
			Prefix: "TaskManager",
		})
		return
	}

	// Stop existing scheduled execution if any
	tm.StopTask(taskName)

	shouldExec, err := tm.shouldExecuteTask(taskName, task.Interval)
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking execution condition for task %s: %v", taskName, err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "TaskManager",
		})
		return
	}

	doneChan := make(chan bool)
	tm.Mutex.Lock()
	tm.Done[taskName] = doneChan
	tm.Mutex.Unlock()

	go func() {
		// Execute immediately if due
		if shouldExec {
			if err := task.Execute(); err != nil {
				tm.logTaskExecution(taskName, "error", err.Error())
				logger.Log(fmt.Sprintf("Task %s execution failed: %v", taskName, err), logger.LogOptions{
					Level:  logger.Error,
					Prefix: "TaskManager",
				})
			} else {
				task.LastRun = time.Now()
				tm.logTaskExecution(taskName, "success", "Task executed successfully")
				logger.Log(fmt.Sprintf("Task %s executed successfully", taskName), logger.LogOptions{
					Level:  logger.Success,
					Prefix: "TaskManager",
				})
			}
		} else {
			// Calculate time until next execution
			var lastLog entities.TaskLog
			var initialDelay time.Duration = task.Interval

			if err := tm.Database.Where("task_name = ?", taskName).Order("executed_at desc").First(&lastLog).Error; err == nil {
				elapsed := time.Since(lastLog.ExecutedAt)
				if elapsed < task.Interval {
					initialDelay = task.Interval - elapsed
				}
			}

			logger.Log(fmt.Sprintf("Task %s will run in %v", taskName, initialDelay), logger.LogOptions{
				Level:  logger.Info,
				Prefix: "TaskManager",
			})

			// Wait for initial delay before first execution
			select {
			case <-time.After(initialDelay):
				if err := task.Execute(); err != nil {
					tm.logTaskExecution(taskName, "error", err.Error())
					logger.Log(fmt.Sprintf("Task %s execution failed: %v", taskName, err), logger.LogOptions{
						Level:  logger.Error,
						Prefix: "TaskManager",
					})
				} else {
					task.LastRun = time.Now()
					tm.logTaskExecution(taskName, "success", "Task executed successfully")
					logger.Log(fmt.Sprintf("Task %s executed successfully", taskName), logger.LogOptions{
						Level:  logger.Success,
						Prefix: "TaskManager",
					})
				}
			case <-doneChan:
				return
			}
		}

		// Skip ticker creation for manual-only tasks (interval = 0)
		if task.Interval == 0 {
			logger.Log(fmt.Sprintf("Task %s is manual-only (no scheduled interval)", taskName), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "TaskManager",
			})
			return
		}

		// Create ticker for subsequent executions
		ticker := time.NewTicker(task.Interval)
		tm.Mutex.Lock()
		tm.Tickers[taskName] = ticker
		tm.Mutex.Unlock()

		// Regular ticker loop
		for {
			select {
			case <-ticker.C:
				if err := task.Execute(); err != nil {
					tm.logTaskExecution(taskName, "error", err.Error())
					logger.Log(fmt.Sprintf("Task %s execution failed: %v", taskName, err), logger.LogOptions{
						Level:  logger.Error,
						Prefix: "TaskManager",
					})
				} else {
					task.LastRun = time.Now()
					tm.logTaskExecution(taskName, "success", "Task executed successfully")
					logger.Log(fmt.Sprintf("Task %s executed successfully", taskName), logger.LogOptions{
						Level:  logger.Success,
						Prefix: "TaskManager",
					})
				}
			case <-doneChan:
				ticker.Stop()
				return
			}
		}
	}()

	logger.Log(fmt.Sprintf("Task %s scheduled with interval %v", taskName, task.Interval), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "TaskManager",
	})
}

func (tm *TaskManager) StopTask(taskName string) {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	if doneChan, exists := tm.Done[taskName]; exists {
		close(doneChan)
		delete(tm.Done, taskName)
		delete(tm.Tickers, taskName)
		logger.Log(fmt.Sprintf("Task %s stopped", taskName), logger.LogOptions{
			Level:  logger.Info,
			Prefix: "TaskManager",
		})
	}
}

// UpdateTaskLastRun manually updates a task's LastRun time
func (tm *TaskManager) UpdateTaskLastRun(taskName string, lastRun time.Time) {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	if task, exists := tm.Tasks[taskName]; exists {
		task.LastRun = lastRun
		tm.Tasks[taskName] = task
		logger.Log(fmt.Sprintf("Updated task %s LastRun: %v", taskName, lastRun), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "TaskManager",
		})
	}
}

func (tm *TaskManager) StartAllTasks() {
	tm.Mutex.Lock()
	var taskNames []string
	for name := range tm.Tasks {
		taskNames = append(taskNames, name)
	}
	tm.Mutex.Unlock()

	for _, taskName := range taskNames {
		tm.StartTask(taskName)
	}
}

func (tm *TaskManager) StopAllTasks() {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	for name, doneChan := range tm.Done {
		close(doneChan)
		delete(tm.Done, name)
		if ticker, exists := tm.Tickers[name]; exists {
			ticker.Stop()
			delete(tm.Tickers, name)
		}
		logger.Log(fmt.Sprintf("Task %s stopped", name), logger.LogOptions{
			Level:  logger.Info,
			Prefix: "TaskManager",
		})
	}
}

func (tm *TaskManager) GetTaskStatus(taskName string) *types.TaskStatus {
	tm.Mutex.Lock()
	task, registered := tm.Tasks[taskName]
	_, running := tm.Tickers[taskName]
	tm.Mutex.Unlock()

	var lastRun, nextRun *time.Time
	var logEntry entities.TaskLog

	if err := tm.Database.Where("task_name = ?", taskName).Order("executed_at desc").First(&logEntry).Error; err == nil {
		lastRun = &logEntry.ExecutedAt
		if logEntry.Status == "error" {
			lastRun = nil
		}

		if task.Interval > 0 {
			next := logEntry.ExecutedAt.Add(task.Interval)
			nextRun = &next
		} else if task.TriggeredBy != "" {
			// For manual tasks triggered by another task, use parent task's next run
			var parentLog entities.TaskLog
			if err := tm.Database.Where("task_name = ?", task.TriggeredBy).Order("executed_at desc").First(&parentLog).Error; err == nil {
				if parentTask, exists := tm.Tasks[task.TriggeredBy]; exists && parentTask.Interval > 0 {
					next := parentLog.ExecutedAt.Add(parentTask.Interval)
					nextRun = &next
				}
			}
		}
	} else if err != gorm.ErrRecordNotFound {
		logger.Log(fmt.Sprintf("Error fetching task log for %s: %v", taskName, err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "TaskManager",
		})
	}

	return &types.TaskStatus{
		Registered: registered,
		Running:    running,
		LastRun:    lastRun,
		NextRun:    nextRun,
	}
}

func (tm *TaskManager) GetAllTaskStatuses() map[string]*types.TaskStatus {
	statuses := make(map[string]*types.TaskStatus)
	tm.Mutex.Lock()
	for name := range tm.Tasks {
		tm.Mutex.Unlock() // temporarily unlock to avoid deadlock / get task status
		statuses[name] = tm.GetTaskStatus(name)
		tm.Mutex.Lock()
	}
	tm.Mutex.Unlock()
	return statuses
}
