package repositories

import (
	"errors"
	"metachan/entities"
	"metachan/utils/logger"

	"gorm.io/gorm/clause"
)

func GetTaskStatus(taskName string) (entities.TaskStatus, error) {
	var taskStatus entities.TaskStatus

	result := DB.Where("task_name = ?", taskName).First(&taskStatus)

	if result.Error != nil {
		return entities.TaskStatus{}, errors.New("task status not found")
	}

	return taskStatus, nil
}

func SetTaskStatus(task *entities.TaskStatus) error {
	result := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "task_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_completed", "last_run_at", "updated_at"}),
	}).Create(task)

	if result.Error != nil {
		logger.Errorf("Task", "Failed to set task status for %s: %v", task.TaskName, result.Error)

		return errors.New("failed to set task status")
	}

	return nil
}

func GetLatestTaskLog(taskName string) (*entities.TaskLog, error) {
	var taskLog entities.TaskLog

	result := DB.Where("task_name = ?", taskName).Order("executed_at desc").First(&taskLog)
	if result.Error != nil {
		return nil, result.Error
	}

	return &taskLog, nil
}

func CreateTaskLog(taskLog *entities.TaskLog) error {
	result := DB.Create(taskLog)
	if result.Error != nil {
		logger.Errorf("Task", "Failed to create task log: %v", result.Error)
		return errors.New("failed to create task log")
	}

	return nil
}

// -- Moved to database/tasks.go --
// import (
// 	"metachan/entities"
// 	"time"
// )

// // MarkTaskComplete marks a task as completed and updates its last run time
// func MarkTaskComplete(taskName string) error {
// 	var taskStatus entities.TaskStatus

// 	result := DB.Where("task_name = ?", taskName).First(&taskStatus)
// 	if result.Error != nil {
// 		// Create new task status if it doesn't exist
// 		taskStatus = entities.TaskStatus{
// 			TaskName:    taskName,
// 			IsCompleted: true,
// 			LastRunAt:   time.Now(),
// 		}
// 		return DB.Create(&taskStatus).Error
// 	}

// 	// Update existing task status
// 	taskStatus.IsCompleted = true
// 	taskStatus.LastRunAt = time.Now()
// 	return DB.Save(&taskStatus).Error
// }

// // IsTaskComplete checks if a task has been completed
// func IsTaskComplete(taskName string) bool {
// 	var taskStatus entities.TaskStatus
// 	result := DB.Where("task_name = ? AND is_completed = ?", taskName, true).First(&taskStatus)
// 	return result.Error == nil
// }

// // GetTaskLastRun returns the last run time for a task
// func GetTaskLastRun(taskName string) *time.Time {
// 	var taskStatus entities.TaskStatus
// 	result := DB.Where("task_name = ?", taskName).First(&taskStatus)
// 	if result.Error != nil {
// 		return nil
// 	}
// 	return &taskStatus.LastRunAt
// }

// // ResetTaskStatus resets a task's completion status (useful for periodic tasks)
// func ResetTaskStatus(taskName string) error {
// 	return DB.Model(&entities.TaskStatus{}).
// 		Where("task_name = ?", taskName).
// 		Update("is_completed", false).Error
// }
