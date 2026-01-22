package migrate

import (
	"regexp"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSplitStatements(t *testing.T) {
	input := "CREATE TABLE a (name TEXT);INSERT INTO a (name) VALUES ('x;y');"
	statements := splitStatements(input)
	require.Len(t, statements, 2)
	require.Equal(t, "CREATE TABLE a (name TEXT)", statements[0])
	require.Equal(t, "INSERT INTO a (name) VALUES ('x;y')", statements[1])
}

func TestLoadMigrationsOrdersByVersion(t *testing.T) {
	fs := fstest.MapFS{
		"002_add.sql":  {Data: []byte("CREATE TABLE b (id INTEGER);")},
		"001_init.sql": {Data: []byte("CREATE TABLE a (id INTEGER);")},
		"README.txt":   {Data: []byte("skip")},
	}
	migrations, err := loadMigrations(fs)
	require.NoError(t, err)
	require.Len(t, migrations, 2)
	require.Equal(t, 1, migrations[0].Version)
	require.Equal(t, 2, migrations[1].Version)
}

func TestSubMigrationsFS(t *testing.T) {
	fs := fstest.MapFS{
		"migrations/postgres/001_init.sql": {Data: []byte("CREATE TABLE a (id INTEGER);")},
	}
	sub, err := SubMigrationsFS(fs, "postgres")
	require.NoError(t, err)

	migrations, err := loadMigrations(sub)
	require.NoError(t, err)
	require.Len(t, migrations, 1)
}

func TestValidateRejectsEmpty(t *testing.T) {
	fs := fstest.MapFS{
		"001_init.sql": {Data: []byte("   ")},
	}
	err := Validate(fs)
	require.Error(t, err)
}

func TestValidateRejectsDuplicateVersions(t *testing.T) {
	fs := fstest.MapFS{
		"001_init.sql":      {Data: []byte("CREATE TABLE a (id INTEGER);")},
		"001_duplicate.sql": {Data: []byte("CREATE TABLE b (id INTEGER);")},
	}
	err := Validate(fs)
	require.Error(t, err)
}

func TestValidateAcceptsMigrations(t *testing.T) {
	fs := fstest.MapFS{
		"001_init.sql": {Data: []byte("CREATE TABLE a (id INTEGER);")},
		"002_next.sql": {Data: []byte("CREATE TABLE b (id INTEGER);")},
	}
	err := Validate(fs)
	require.NoError(t, err)
}

func TestApplySkipsAppliedMigrations(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(
		"CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())",
	)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT version FROM schema_migrations ORDER BY version",
	)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(1))

	fs := fstest.MapFS{
		"001_init.sql": {Data: []byte("CREATE TABLE a (id INTEGER)")},
		"002_add.sql":  {Data: []byte("CREATE TABLE b (id INTEGER)")},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE b (id INTEGER)")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (version) VALUES ($1)")).WithArgs(2).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = Apply(db, "postgres", fs)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyUsesSQLitePlaceholders(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(
		"CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)",
	)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT version FROM schema_migrations ORDER BY version",
	)).WillReturnRows(sqlmock.NewRows([]string{"version"}))

	fs := fstest.MapFS{
		"001_init.sql": {Data: []byte("CREATE TABLE a (id INTEGER)")},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE a (id INTEGER)")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (version) VALUES (?)")).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = Apply(db, "sqlite", fs)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
