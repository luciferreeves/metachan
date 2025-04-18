package types

type DatabaseDriver string

const (
	SQLite    DatabaseDriver = "sqlite"
	MySQL     DatabaseDriver = "mysql"
	Postgres  DatabaseDriver = "postgres"
	SQLServer DatabaseDriver = "sqlserver"
)

type TMDBConfig struct {
	APIKey          string
	ReadAccessToken string
}

type ServerConfig struct {
	DatabaseDriver DatabaseDriver
	DataSourceName string
	Port           int
	TMDB           TMDBConfig
}
