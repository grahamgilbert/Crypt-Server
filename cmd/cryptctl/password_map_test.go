package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadPasswordMap(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "passwords.csv")
	err := os.WriteFile(path, []byte("username_or_email,password,must_reset_password\nadmin,Str0ng!Passw0rd,false\n"), 0o600)
	require.NoError(t, err)

	entries, err := loadPasswordMap(path)
	require.NoError(t, err)
	entry, ok := entries["admin"]
	require.True(t, ok)
	require.False(t, entry.MustResetPassword)
	require.NotEmpty(t, entry.PasswordHash)
}

func TestLoadPasswordMapErrors(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "passwords.csv")
	err := os.WriteFile(path, []byte("user, ,\n"), 0o600)
	require.NoError(t, err)

	_, err = loadPasswordMap(path)
	require.Error(t, err)
}
