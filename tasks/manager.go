package tasks

import (
	"fmt"
	"metachan/entities"
	"metachan/repositories"
	"metachan/types"
	"metachan/utils/logger"
	"sync"
	"time"

	"gorm.io/gorm"
)

type TaskManager struct {
	Tasks   map[string]types.Task
	Tickers map[string]*time.Ticker
	Done    map[string]chan bool
	Mutex   sync.Mutex
}

func (tm *TaskManager) RegisterTask(task types.Task) error {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	if _, exists := tm.Tasks[task.Name]; exists {
		return fmt.Errorf("task %s already registered", task.Name)
	}

	tm.Tasks[task.Name] = task
	logger.Infof("TaskManager", "Task %s registered", task.Name)

	return nil
}

func (tm *TaskManager) shouldExecuteTask(taskName string, interval time.Duration) (bool, error) {
	lastLog, err := repositories.GetLatestTaskLog(taskName)
	if err != nil {
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

	if err := repositories.CreateTaskLog(&logEntry); err != nil {
		logger.Warnf("TaskManager", "Failed to log task execution for %s: %v", taskName, err)
	}
}

func (tm *TaskManager) StartTask(taskName string) {
	tm.Mutex.Lock()
	task, exists := tm.Tasks[taskName]
	tm.Mutex.Unlock()
	if !exists {
		logger.Warnf("TaskManager", "Task %s not found", taskName)
		return
	}

	// Stop existing scheduled execution if any
	tm.StopTask(taskName)

	shouldExec, err := tm.shouldExecuteTask(taskName, task.Interval)
	if err != nil {
		logger.Errorf("TaskManager", "Error checking execution condition for task %s: %v", taskName, err)
		return
	}

	doneChan := make(chan bool)
	tm.Mutex.Lock()
	tm.Done[taskName] = doneChan
	tm.Mutex.Unlock()

	go func() {
		// Execute immediately if due
		if shouldExec {
			// Check dependencies before executing
			if !tm.checkDependencies(task) {
				logger.Warnf("TaskManager", "Task %s dependencies not met, skipping execution", taskName)
			} else if err := task.Execute(); err != nil {
				tm.logTaskExecution(taskName, "error", err.Error())
				logger.Errorf("TaskManager", "Task %s execution failed: %v", taskName, err)
			} else {
				task.LastRun = time.Now()
				tm.logTaskExecution(taskName, "success", "Task executed successfully")
				logger.Successf("TaskManager", "Task %s executed successfully", taskName)
			}
		} else {
			// Calculate time until next execution
			var initialDelay time.Duration = task.Interval

			if lastLog, err := repositories.GetLatestTaskLog(taskName); err == nil {
				elapsed := time.Since(lastLog.ExecutedAt)
				if elapsed < task.Interval {
					initialDelay = task.Interval - elapsed
				}
			}

			logger.Infof("TaskManager", "Task %s will run in %v", taskName, initialDelay)

			// Wait for initial delay before first execution
			select {
			case <-time.After(initialDelay):
				// Check dependencies before executing
				if !tm.checkDependencies(task) {
					logger.Warnf("TaskManager", "Task %s dependencies not met, skipping execution", taskName)
				} else if err := task.Execute(); err != nil {
					tm.logTaskExecution(taskName, "error", err.Error())
					logger.Errorf("TaskManager", "Task %s execution failed: %v", taskName, err)
				} else {
					task.LastRun = time.Now()
					tm.logTaskExecution(taskName, "success", "Task executed successfully")
					logger.Successf("TaskManager", "Task %s executed successfully", taskName)
				}
			case <-doneChan:
				return
			}
		}

		// Skip ticker creation for manual-only tasks (interval = 0)
		if task.Interval == 0 {
			logger.Debugf("TaskManager", "Task %s is manual-only (no scheduled interval)", taskName)
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
				// Check dependencies before executing
				if !tm.checkDependencies(task) {
					logger.Warnf("TaskManager", "Task %s dependencies not met, skipping execution", taskName)
				} else if err := task.Execute(); err != nil {
					tm.logTaskExecution(taskName, "error", err.Error())
					logger.Errorf("TaskManager", "Task %s execution failed: %v", taskName, err)
				} else {
					task.LastRun = time.Now()
					tm.logTaskExecution(taskName, "success", "Task executed successfully")
					logger.Successf("TaskManager", "Task %s executed successfully", taskName)
				}
			case <-doneChan:
				ticker.Stop()
				return
			}
		}
	}()

	logger.Infof("TaskManager", "Task %s scheduled with interval %v", taskName, task.Interval)
}

func (tm *TaskManager) StopTask(taskName string) {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	if doneChan, exists := tm.Done[taskName]; exists {
		close(doneChan)
		delete(tm.Done, taskName)
		delete(tm.Tickers, taskName)
		logger.Infof("TaskManager", "Task %s stopped", taskName)
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
		logger.Infof("TaskManager", "Task %s stopped", name)
	}
}

// checkDependencies verifies all task dependencies are complete
func (tm *TaskManager) checkDependencies(task types.Task) bool {
	if len(task.Dependencies) == 0 {
		return true
	}

	for _, depName := range task.Dependencies {
		taskStatus, err := repositories.GetTaskStatus(depName)
		if err != nil || !taskStatus.IsCompleted {
			logger.Debugf("TaskManager", "Dependency %s not completed for task %s", depName, task.Name)
			return false
		}
	}

	return true
}

func (tm *TaskManager) GetTaskStatus(taskName string) *types.TaskStatus {
	tm.Mutex.Lock()
	task, registered := tm.Tasks[taskName]
	_, running := tm.Tickers[taskName]
	tm.Mutex.Unlock()

	var lastRun, nextRun *time.Time

	if logEntry, err := repositories.GetLatestTaskLog(taskName); err == nil {
		lastRun = &logEntry.ExecutedAt
		if logEntry.Status == "error" {
			lastRun = nil
		}

		if task.Interval > 0 {
			next := logEntry.ExecutedAt.Add(task.Interval)
			nextRun = &next
		}
	} else if err != gorm.ErrRecordNotFound {
		logger.Errorf("TaskManager", "Error fetching task log for %s: %v", taskName, err)
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
