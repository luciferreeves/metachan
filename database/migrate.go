package database

import (
	"fmt"
	"metachan/entities"
	"metachan/utils/logger"
)

func AutoMigrate() {
	err := DB.AutoMigrate(
		&entities.TaskLog{},
		&entities.AnimeMapping{},
	)
	if err != nil {
		logger.Log(fmt.Sprintf("Error during auto migration: %v", err), logger.LogOptions{
			Prefix: "Database",
			Level:  logger.Error,
			Fatal:  true,
		})
	} else {
		logger.Log("Auto migration completed successfully", logger.LogOptions{
			Prefix: "Database",
			Level:  logger.Success,
		})
	}
}
