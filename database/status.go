package database

import (
	"context"
	"fmt"
	"metachan/types"
	"metachan/utils/logger"
	"time"
)

func DatabaseConnectionStatus() bool {
	if DB == nil {
		return false
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log(fmt.Sprintf("Unable to get SQL DB: %v", err), types.LogOptions{
			Prefix: "Database",
			Level:  types.Error,
		})
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sqlDB.PingContext(ctx)
	if err != nil {
		logger.Log(fmt.Sprintf("Database connection error: %v", err), types.LogOptions{
			Prefix: "Database",
			Level:  types.Error,
		})
		return false
	}
	return true
}
