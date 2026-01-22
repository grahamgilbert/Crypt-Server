package main

import (
	"fmt"
	"strings"

	"crypt-server/internal/app"
	"crypt-server/internal/store"
)

func createFirstAdmin(dataStore store.Store, username, password string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("admin username is required")
	}
	if password == "" {
		return fmt.Errorf("admin password is required")
	}
	users, err := dataStore.ListUsers()
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}
	if len(users) > 0 {
		return fmt.Errorf("cannot create first admin: users already exist")
	}
	hash, err := app.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = dataStore.AddUser(username, hash, true, true, true, false, "local")
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}

func resetUserPassword(dataStore store.Store, username, password string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}
	user, err := dataStore.GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	hash, err := app.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = dataStore.UpdateUserPassword(user.ID, hash, false)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
