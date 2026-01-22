package main

import (
	"fmt"
	"os"
	"strings"
)

type databaseConfig struct {
	driver string
	dsn    string
}

func loadDatabaseConfig() (databaseConfig, error) {
	postgresURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	sqlitePath := strings.TrimSpace(os.Getenv("SQLITE_PATH"))

	if postgresURL != "" && sqlitePath != "" {
		return databaseConfig{}, fmt.Errorf("set only one of DATABASE_URL or SQLITE_PATH")
	}
	if postgresURL != "" {
		return databaseConfig{driver: "postgres", dsn: postgresURL}, nil
	}
	if sqlitePath != "" {
		if isSQLiteMemory(sqlitePath) {
			return databaseConfig{}, fmt.Errorf("SQLITE_PATH must point to a file, not an in-memory database")
		}
		return databaseConfig{driver: "sqlite", dsn: sqlitePath}, nil
	}
	return databaseConfig{}, fmt.Errorf("DATABASE_URL or SQLITE_PATH is required")
}

func isSQLiteMemory(path string) bool {
	cleaned := strings.ToLower(strings.TrimSpace(path))
	if cleaned == ":memory:" {
		return true
	}
	return strings.Contains(cleaned, "mode=memory")
}
