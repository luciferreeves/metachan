package database

import (
	"fmt"
	"metachan/config"
	"metachan/types"
	"metachan/utils/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func init() {
	var dialector gorm.Dialector

	switch config.Config.DatabaseDriver {
	case types.Postgres:
		dialector = postgres.Open(config.Config.DataSourceName)
	case types.SQLite:
		dialector = sqlite.Open(config.Config.DataSourceName)
	case types.MySQL:
		dialector = mysql.Open(config.Config.DataSourceName)
	case types.SQLServer:
		dialector = sqlserver.Open(config.Config.DataSourceName)
	default:
		logger.Log(fmt.Sprintf("Invalid database driver: %s", config.Config.DatabaseDriver), logger.LogOptions{
			Prefix: "Database",
			Level:  logger.Error,
			Fatal:  true,
		})
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		logger.Log(fmt.Sprintf("Error connecting to database: %v", err), logger.LogOptions{
			Prefix: "Database",
			Level:  logger.Error,
			Fatal:  true,
		})
	} else {
		logger.Log("Database connection established successfully", logger.LogOptions{
			Prefix: "Database",
			Level:  logger.Success,
		})
	}
}
