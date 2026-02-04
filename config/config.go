package config

import (
	"metachan/utils/env"
	"metachan/utils/logger"

	"github.com/joho/godotenv"
)

var (
	Server   server
	Database database
	Sync     sync
	API      api
)

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Infof("Config", "No .env file found. Environment variables will be used directly.")
	}

	if err := env.Parse(&Server); err != nil {
		logger.Fatalf("Config", "Failed to parse server config: %v", err)
	}

	if err := env.Parse(&Database); err != nil {
		logger.Fatalf("Config", "Failed to parse database config: %v", err)
	}

	if err := env.Parse(&Sync); err != nil {
		logger.Fatalf("Config", "Failed to parse sync config: %v", err)
	}

	if err := env.Parse(&API); err != nil {
		logger.Fatalf("Config", "Failed to parse API config: %v", err)
	}

	if Server.Debug {
		logger.SetDebug(true)
	}

	if err := verifyConfig(); err != nil {
		logger.Fatalf("Config", "Configuration verification failed: %v", err)
	}

	logger.Successf("Config", "Configuration loaded successfully")
}
