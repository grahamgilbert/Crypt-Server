package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadKeyPriority(t *testing.T) {
	os.Setenv("FIELD_ENCRYPTION_KEY", "env-key")
	t.Cleanup(func() { os.Unsetenv("FIELD_ENCRYPTION_KEY") })

	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "key.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("file-key"), 0o600))

	value, err := loadKey("flag-key", filePath, "FIELD_ENCRYPTION_KEY")
	require.NoError(t, err)
	require.Equal(t, "flag-key", value)

	value, err = loadKey("", filePath, "FIELD_ENCRYPTION_KEY")
	require.NoError(t, err)
	require.Equal(t, "file-key", value)

	value, err = loadKey("", "", "FIELD_ENCRYPTION_KEY")
	require.NoError(t, err)
	require.Equal(t, "env-key", value)
}

func TestLoadKeyMissing(t *testing.T) {
	os.Unsetenv("FIELD_ENCRYPTION_KEY")
	_, err := loadKey("", "", "FIELD_ENCRYPTION_KEY")
	require.Error(t, err)
}
