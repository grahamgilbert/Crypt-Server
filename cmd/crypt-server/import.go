package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"crypt-server/internal/fixture"
	"crypt-server/internal/store"
)

var ErrDatabaseNotEmpty = errors.New("database is not empty; import only allowed on empty database")

func importFixture(st store.Store, fixturePath string) error {
	// Check if database is empty first
	isEmpty, err := st.IsEmpty()
	if err != nil {
		return fmt.Errorf("check database: %w", err)
	}
	if !isEmpty {
		return ErrDatabaseNotEmpty
	}

	// Read and parse the fixture file
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		return fmt.Errorf("read fixture file: %w", err)
	}

	var migration fixture.MigrationOutput
	if err := json.Unmarshal(data, &migration); err != nil {
		return fmt.Errorf("parse fixture: %w", err)
	}

	// Import computers first (secrets depend on them)
	for _, c := range migration.Computers {
		lastCheckin, err := parseDateTime(c.LastCheckin)
		if err != nil {
			return fmt.Errorf("parse computer %d last_checkin: %w", c.ID, err)
		}
		if err := st.ImportComputer(c.ID, c.Serial, c.Username, c.ComputerName, lastCheckin); err != nil {
			return fmt.Errorf("import computer %d: %w", c.ID, err)
		}
	}

	// Import secrets (requests depend on them)
	for _, s := range migration.Secrets {
		dateEscrowed, err := parseDateTime(s.DateEscrowed)
		if err != nil {
			return fmt.Errorf("parse secret %d date_escrowed: %w", s.ID, err)
		}
		if err := st.ImportSecret(s.ID, s.ComputerID, s.SecretType, s.Secret, dateEscrowed, s.RotationRequired); err != nil {
			return fmt.Errorf("import secret %d: %w", s.ID, err)
		}
	}

	// Import users
	for _, u := range migration.Users {
		if err := st.ImportUser(u.ID, u.Username, u.PasswordHash, u.IsStaff, u.CanApprove, u.LocalLoginEnabled, u.MustResetPassword, u.AuthSource); err != nil {
			return fmt.Errorf("import user %d: %w", u.ID, err)
		}
	}

	// Import requests
	for _, r := range migration.Requests {
		dateRequested, err := parseDateTime(r.DateRequested)
		if err != nil {
			return fmt.Errorf("parse request %d date_requested: %w", r.ID, err)
		}
		var dateApproved *time.Time
		if r.DateApproved != "" {
			da, err := parseDateTime(r.DateApproved)
			if err != nil {
				return fmt.Errorf("parse request %d date_approved: %w", r.ID, err)
			}
			dateApproved = &da
		}
		if err := st.ImportRequest(r.ID, r.SecretID, r.RequestingUser, r.Approved, r.AuthUser, r.ReasonForRequest, r.ReasonForApproval, dateRequested, dateApproved, r.Current); err != nil {
			return fmt.Errorf("import request %d: %w", r.ID, err)
		}
	}

	return nil
}

// parseDateTime parses a datetime string from Django fixtures.
// Supports formats: "2006-01-02T15:04:05Z", "2006-01-02T15:04:05.000Z", "2006-01-02 15:04:05"
func parseDateTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", value)
}
