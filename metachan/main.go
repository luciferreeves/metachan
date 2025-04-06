package main

import (
	"fmt"
	"metachan/config"
	"metachan/database"
	"metachan/tasks"
	"metachan/types"
	"metachan/utils/logger"
)

func main() {
	database.AutoMigrate()

	tasks.GlobalTaskManager.StartAllTasks()

	logger.Log(fmt.Sprintf("Server started on port %d. Database Driver is: %s. Configured DSN is: %s", config.Config.Port, config.Config.DatabaseDriver, config.Config.DataSourceName), types.LogOptions{
		Prefix: "Main",
		Level:  types.Info,
		Fatal:  false,
	})

	select {}
	// Keep the main function running
}
