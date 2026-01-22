package main

import (
	"fmt"
	"io"
	"io/fs"
	"strings"

	"crypt-server/internal/migrate"
)

func migrationDrivers(driver string) ([]string, error) {
	switch strings.TrimSpace(driver) {
	case "":
		return []string{"postgres", "sqlite"}, nil
	case "postgres":
		return []string{"postgres"}, nil
	case "sqlite":
		return []string{"sqlite"}, nil
	default:
		return nil, fmt.Errorf("unsupported migrations driver: %s", driver)
	}
}

func runMigrationCommand(w io.Writer, fsys fs.FS, driver string, validate, print bool) error {
	drivers, err := migrationDrivers(driver)
	if err != nil {
		return err
	}
	for _, dbDriver := range drivers {
		sub, err := migrate.SubMigrationsFS(fsys, dbDriver)
		if err != nil {
			return fmt.Errorf("load %s migrations: %w", dbDriver, err)
		}
		if validate {
			if err := migrate.Validate(sub); err != nil {
				return fmt.Errorf("%s migrations invalid: %w", dbDriver, err)
			}
		}
		if print {
			if err := printMigrations(w, sub, dbDriver); err != nil {
				return err
			}
		}
	}
	return nil
}

func printMigrations(w io.Writer, fsys fs.FS, driver string) error {
	migrations, err := migrate.List(fsys)
	if err != nil {
		return fmt.Errorf("list %s migrations: %w", driver, err)
	}
	if len(migrations) == 0 {
		return fmt.Errorf("no %s migrations found", driver)
	}
	fmt.Fprintf(w, "== %s ==\n", driver)
	for _, migration := range migrations {
		fmt.Fprintf(w, "-- %03d %s\n", migration.Version, migration.Name)
		fmt.Fprintf(w, "%s\n", strings.TrimRight(migration.SQL, "\n"))
		fmt.Fprintln(w)
	}
	return nil
}
