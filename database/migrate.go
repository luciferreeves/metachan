package database

import (
	"fmt"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
)

func AutoMigrate() {
	err := DB.AutoMigrate(
		&entities.TaskLog{},
		&entities.AnimeMapping{},
	)
	if err != nil {
		logger.Log(fmt.Sprintf("Error during auto migration: %v", err), types.LogOptions{
			Prefix: "Database",
			Level:  types.Error,
			Fatal:  true,
		})
	} else {
		logger.Log("Auto migration completed successfully", types.LogOptions{
			Prefix: "Database",
			Level:  types.Success,
		})
	}
}
