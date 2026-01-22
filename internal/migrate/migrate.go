package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

var migrationPattern = regexp.MustCompile(`^(\d+)_.*\.sql$`)

func Apply(db *sql.DB, driver string, fsys fs.FS) error {
	if err := ensureSchemaMigrations(db, driver); err != nil {
		return err
	}
	migrations, err := loadMigrations(fsys)
	if err != nil {
		return err
	}
	applied, err := loadAppliedVersions(db)
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}
		if err := applyMigration(db, driver, migration); err != nil {
			return err
		}
	}
	return nil
}

func List(fsys fs.FS) ([]Migration, error) {
	return loadMigrations(fsys)
}

func Validate(fsys fs.FS) error {
	migrations, err := loadMigrations(fsys)
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		return fmt.Errorf("no migrations found")
	}
	versions := make(map[int]struct{})
	for _, migration := range migrations {
		if strings.TrimSpace(migration.SQL) == "" {
			return fmt.Errorf("migration %s is empty", migration.Name)
		}
		if _, exists := versions[migration.Version]; exists {
			return fmt.Errorf("duplicate migration version %d", migration.Version)
		}
		versions[migration.Version] = struct{}{}
	}
	return nil
}

func ensureSchemaMigrations(db *sql.DB, driver string) error {
	stmt, err := schemaMigrationsSQL(driver)
	if err != nil {
		return err
	}
	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return nil
}

func schemaMigrationsSQL(driver string) (string, error) {
	switch driver {
	case "postgres":
		return "CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())", nil
	case "sqlite":
		return "CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)", nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", driver)
	}
}

func loadMigrations(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	migrations := make([]Migration, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		matches := migrationPattern.FindStringSubmatch(name)
		if matches == nil {
			continue
		}
		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version %s: %w", name, err)
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}
		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(data),
		})
	}
	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].Version == migrations[j].Version {
			return migrations[i].Name < migrations[j].Name
		}
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

func loadAppliedVersions(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func applyMigration(db *sql.DB, driver string, migration Migration) error {
	statements := splitStatements(migration.SQL)
	if len(statements) == 0 {
		return fmt.Errorf("migration %s is empty", migration.Name)
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.Name, err)
	}
	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("rollback migration %s: %v", migration.Name, rollbackErr)
			}
			return fmt.Errorf("apply migration %s: %w", migration.Name, err)
		}
	}
	insertSQL, err := insertMigrationSQL(driver)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec(insertSQL, migration.Version); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("rollback migration %s: %v", migration.Name, rollbackErr)
		}
		return fmt.Errorf("record migration %s: %w", migration.Name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.Name, err)
	}
	return nil
}

func insertMigrationSQL(driver string) (string, error) {
	switch driver {
	case "postgres":
		return "INSERT INTO schema_migrations (version) VALUES ($1)", nil
	case "sqlite":
		return "INSERT INTO schema_migrations (version) VALUES (?)", nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", driver)
	}
}

func splitStatements(sqlText string) []string {
	trimmed := strings.TrimSpace(sqlText)
	if trimmed == "" {
		return nil
	}
	statements := make([]string, 0)
	var current strings.Builder
	inSingle := false
	inDouble := false
	var prev rune
	for _, ch := range sqlText {
		if ch == '\'' && !inDouble && prev != '\\' {
			inSingle = !inSingle
		} else if ch == '"' && !inSingle && prev != '\\' {
			inDouble = !inDouble
		}
		if ch == ';' && !inSingle && !inDouble {
			statement := strings.TrimSpace(current.String())
			if statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
			prev = ch
			continue
		}
		current.WriteRune(ch)
		prev = ch
	}
	statement := strings.TrimSpace(current.String())
	if statement != "" {
		statements = append(statements, statement)
	}
	return statements
}

func SubMigrationsFS(fsys fs.FS, driver string) (fs.FS, error) {
	return fs.Sub(fsys, filepath.Join("migrations", driver))
}
