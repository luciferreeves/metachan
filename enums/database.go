package enums

type DatabaseDriver string

const (
	SQLite    DatabaseDriver = "sqlite"
	MySQL     DatabaseDriver = "mysql"
	Postgres  DatabaseDriver = "postgres"
	SQLServer DatabaseDriver = "sqlserver"
)
