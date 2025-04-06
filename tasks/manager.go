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
	logger.Log(fmt.Sprintf("Task %s registered", task.Name), types.LogOptions{
		Level:  types.Info,
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
		logger.Log(fmt.Sprintf("Failed to log task execution for %s: %v", taskName, err), types.LogOptions{
			Level:  types.Warn,
			Prefix: "TaskManager",
		})
	}
}

func (tm *TaskManager) StartTask(taskName string) {
	tm.Mutex.Lock()
	task, exists := tm.Tasks[taskName]
	tm.Mutex.Unlock()
	if !exists {
		logger.Log(fmt.Sprintf("Task %s not found", taskName), types.LogOptions{
			Level:  types.Warn,
			Prefix: "TaskManager",
		})
		return
	}

	shouldExec, err := tm.shouldExecuteTask(taskName, task.Interval)
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking execution condition for task %s: %v", taskName, err), types.LogOptions{
			Level:  types.Error,
			Prefix: "TaskManager",
		})
		return
	}

	if !shouldExec {
		logger.Log(fmt.Sprintf("Task %s skipped execution - not due yet", taskName), types.LogOptions{
			Level:  types.Info,
			Prefix: "TaskManager",
		})
		return
	}

	// Stop existing scheduled execution if any
	tm.StopTask(taskName)

	// Execute the task immediately
	go func() {
		if err := task.Execute(); err != nil {
			tm.logTaskExecution(taskName, "error", err.Error())
			logger.Log(fmt.Sprintf("Task %s execution failed: %v", taskName, err), types.LogOptions{
				Level:  types.Error,
				Prefix: "TaskManager",
			})
		} else {
			task.LastRun = time.Now()
			tm.logTaskExecution(taskName, "success", "Task executed successfully")
			logger.Log(fmt.Sprintf("Task %s executed successfully", taskName), types.LogOptions{
				Level:  types.Success,
				Prefix: "TaskManager",
			})
		}
	}()

	// Schedule subsequent executions
	ticker := time.NewTicker(task.Interval)
	doneChan := make(chan bool)
	tm.Mutex.Lock()
	tm.Tickers[taskName] = ticker
	tm.Done[taskName] = doneChan
	tm.Mutex.Unlock()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := task.Execute(); err != nil {
					tm.logTaskExecution(taskName, "error", err.Error())
					logger.Log(fmt.Sprintf("Task %s execution failed: %v", taskName, err), types.LogOptions{
						Level:  types.Error,
						Prefix: "TaskManager",
					})
				} else {
					task.LastRun = time.Now()
					tm.logTaskExecution(taskName, "success", "Task executed successfully")
					logger.Log(fmt.Sprintf("Task %s executed successfully", taskName), types.LogOptions{
						Level:  types.Success,
						Prefix: "TaskManager",
					})
				}
			case <-doneChan:
				ticker.Stop()
				return
			}
		}
	}()

	logger.Log(fmt.Sprintf("Task %s scheduled with interval %v", taskName, task.Interval), types.LogOptions{
		Level:  types.Info,
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
		logger.Log(fmt.Sprintf("Task %s stopped", taskName), types.LogOptions{
			Level:  types.Info,
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
		logger.Log(fmt.Sprintf("Task %s stopped", name), types.LogOptions{
			Level:  types.Info,
			Prefix: "TaskManager",
		})
	}
}
