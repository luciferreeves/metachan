package database

import (
	"metachan/config"
	"metachan/enums"
	"metachan/utils/logger"
	"time"

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

	switch enums.DatabaseDriver(config.Database.Driver) {
	case enums.SQLite:
		dialector = sqlite.Open(config.Database.DSN)
	case enums.MySQL:
		dialector = mysql.Open(config.Database.DSN)
	case enums.Postgres:
		dialector = postgres.Open(config.Database.DSN)
	case enums.SQLServer:
		dialector = sqlserver.Open(config.Database.DSN)
	default:
		logger.Fatalf("Database", "Invalid database driver: %s", config.Database.Driver)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})

	if err != nil {
		logger.Fatalf("Database", "Error connecting to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Fatalf("Database", "Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Successf("Database", "Database connection established successfully")

	migrate()
}
