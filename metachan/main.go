package main

import (
	"fmt"
	"metachan/config"
	"metachan/types"
	"metachan/utils/logger"
)

func main() {
	logger.Log(fmt.Sprintf("Server started on port %d. Database Driver is: %s. Configured DSN is: %s", config.Config.Port, config.Config.DatabaseDriver, config.Config.DataSourceName), types.LogOptions{
		Timestamp: true,
		Prefix:    "Main",
		Level:     types.Info,
		Fatal:     false,
	})
}
