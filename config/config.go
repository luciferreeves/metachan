package config

import (
	"metachan/types"
	"metachan/utils/logger"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var Config *types.ServerConfig

func init() {
	logOptions := logger.LogOptions{
		Prefix: "Config",
		Level:  logger.Error,
		Fatal:  true,
	}

	godotenv.Load()

	Config = &types.ServerConfig{
		DatabaseDriver: types.DatabaseDriver(getEnv("DB_DRIVER")),
		DataSourceName: getEnv("DSN"),
		Port:           getIntEnv("PORT"),
		TMDB: types.TMDBConfig{
			APIKey:          getEnv("TMDB_API_KEY"),
			ReadAccessToken: getEnv("TMDB_READ_ACCESS_TOKEN"),
		},
	}

	switch Config.DatabaseDriver {
	case types.SQLite, types.MySQL, types.Postgres, types.SQLServer:
	default:
		logger.Log("Invalid database driver or database driver not set. Valid options are: sqlite, mysql, postgres, sqlserver", logOptions)
	}

	if Config.DataSourceName == "" {
		logger.Log("Invalid data source name or data source name not set", logOptions)
	}

	if Config.Port == 0 {
		logger.Log("Invalid port or port not set", logOptions)
	}

	if Config.TMDB.APIKey == "" {
		logger.Log("Invalid TMDB API key or TMDB API key not set", logOptions)
	}

	if Config.TMDB.ReadAccessToken == "" {
		logger.Log("Invalid TMDB read access token or TMDB read access token not set", logOptions)
	}

	logOptions.Level = logger.Success
	logOptions.Fatal = false
	logger.Log("Config initialized successfully", logOptions)
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return ""
	}
	return strings.TrimSpace(value)
}

func getIntEnv(key string) int {
	value := getEnv(key)
	if value == "" {
		return 0
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return i
}
