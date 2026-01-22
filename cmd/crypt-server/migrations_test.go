package main

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestMigrationDrivers(t *testing.T) {
	drivers, err := migrationDrivers("")
	require.NoError(t, err)
	require.Equal(t, []string{"postgres", "sqlite"}, drivers)

	drivers, err = migrationDrivers("postgres")
	require.NoError(t, err)
	require.Equal(t, []string{"postgres"}, drivers)

	drivers, err = migrationDrivers("sqlite")
	require.NoError(t, err)
	require.Equal(t, []string{"sqlite"}, drivers)

	_, err = migrationDrivers("mysql")
	require.Error(t, err)
}

func TestRunMigrationCommandPrints(t *testing.T) {
	fsys := fstest.MapFS{
		"migrations/postgres/001_init.sql": {Data: []byte("CREATE TABLE a (id INTEGER);")},
		"migrations/sqlite/001_init.sql":   {Data: []byte("CREATE TABLE a (id INTEGER);")},
	}
	var buf bytes.Buffer

	err := runMigrationCommand(&buf, fsys, "postgres", false, true)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "== postgres ==")
	require.Contains(t, buf.String(), "001_init.sql")
}

func TestRunMigrationCommandValidates(t *testing.T) {
	fsys := fstest.MapFS{
		"migrations/sqlite/001_init.sql": {Data: []byte("CREATE TABLE a (id INTEGER);")},
	}

	var buf bytes.Buffer
	err := runMigrationCommand(&buf, fsys, "sqlite", true, false)
	require.NoError(t, err)
	require.Equal(t, "", buf.String())
}
