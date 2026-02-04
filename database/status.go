package database

import (
	"context"
	"metachan/utils/logger"
	"time"
)

func GetConnectionStatus() bool {
	if DB == nil {
		return false
	}

	instance, err := DB.DB()

	if err != nil {
		logger.Errorf("Database", "Failed to get DB instance: %v", err)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = instance.PingContext(ctx)
	if err != nil {
		logger.Errorf("Database", "Database connection error: %v", err)
		return false
	}
	return true
}
