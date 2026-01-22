package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type passwordMapEntry struct {
	PasswordHash      string
	MustResetPassword bool
}

func loadPasswordMap(path string) (map[string]passwordMapEntry, error) {
	if path == "" {
		return map[string]passwordMapEntry{}, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	out := make(map[string]passwordMapEntry)
	for i, record := range records {
		if len(record) == 0 {
			continue
		}
		if i == 0 && isPasswordMapHeader(record) {
			continue
		}
		if len(record) < 2 || len(record) > 3 {
			return nil, fmt.Errorf("invalid password map record %d", i+1)
		}
		key := strings.ToLower(strings.TrimSpace(record[0]))
		if key == "" {
			return nil, fmt.Errorf("invalid password map record %d", i+1)
		}
		password := strings.TrimSpace(record[1])
		if password == "" {
			return nil, fmt.Errorf("password missing for %s", key)
		}
		mustReset := false
		if len(record) == 3 && strings.TrimSpace(record[2]) != "" {
			value, err := strconv.ParseBool(strings.TrimSpace(record[2]))
			if err != nil {
				return nil, fmt.Errorf("invalid must_reset_password value for %s", key)
			}
			mustReset = value
		}
		hash, err := hashPasswordForExport(password)
		if err != nil {
			return nil, err
		}
		if _, exists := out[key]; exists {
			return nil, fmt.Errorf("duplicate password map entry for %s", key)
		}
		out[key] = passwordMapEntry{
			PasswordHash:      hash,
			MustResetPassword: mustReset,
		}
	}
	return out, nil
}

func isPasswordMapHeader(record []string) bool {
	if len(record) < 2 {
		return false
	}
	first := strings.ToLower(strings.TrimSpace(record[0]))
	return first == "username_or_email"
}

func hashPasswordForExport(plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("password is required")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(plaintext), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$%d$%d$%d$%s$%s",
		1,
		64*1024,
		4,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}
