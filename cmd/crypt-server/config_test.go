package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDatabaseConfigPostgres(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SQLITE_PATH", "")

	cfg, err := loadDatabaseConfig()
	require.NoError(t, err)
	require.Equal(t, "postgres", cfg.driver)
	require.Equal(t, "postgres://example", cfg.dsn)
}

func TestLoadDatabaseConfigSQLite(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SQLITE_PATH", "data/crypt.db")

	cfg, err := loadDatabaseConfig()
	require.NoError(t, err)
	require.Equal(t, "sqlite", cfg.driver)
	require.Equal(t, "data/crypt.db", cfg.dsn)
}

func TestLoadDatabaseConfigRejectsBoth(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SQLITE_PATH", "data/crypt.db")

	_, err := loadDatabaseConfig()
	require.Error(t, err)
}

func TestLoadDatabaseConfigRejectsMissing(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SQLITE_PATH", "")

	_, err := loadDatabaseConfig()
	require.Error(t, err)
}

func TestLoadDatabaseConfigRejectsSQLiteMemory(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SQLITE_PATH", ":memory:")

	_, err := loadDatabaseConfig()
	require.Error(t, err)
}

func TestLoadDatabaseConfigRejectsSQLiteMemoryMode(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SQLITE_PATH", "file:crypt.db?mode=memory&cache=shared")

	_, err := loadDatabaseConfig()
	require.Error(t, err)
}
