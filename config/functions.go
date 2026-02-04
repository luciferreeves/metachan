package config

import (
	"fmt"
	"metachan/enums"
)

func verifyConfig() error {
	if Server.Port <= 0 || Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", Server.Port)
	}

	if !verifyDatabaseDriver(enums.DatabaseDriver(Database.Driver)) {
		return fmt.Errorf("invalid database driver: %s", Database.Driver)
	}

	if Database.DSN == "" {
		return fmt.Errorf("data source name (DSN) cannot be empty")
	}

	if API.TMDBKey == "" {
		return fmt.Errorf("TMDB API key cannot be empty")
	}

	if API.TMDBReadToken == "" {
		return fmt.Errorf("TMDB read access token cannot be empty")
	}

	if API.TVDBKey == "" {
		return fmt.Errorf("TVDB API key cannot be empty")
	}

	return nil
}

func verifyDatabaseDriver(driver enums.DatabaseDriver) bool {
	switch driver {
	case enums.SQLite, enums.MySQL, enums.Postgres, enums.SQLServer:
		return true
	default:
		return false
	}
}
