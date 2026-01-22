package main

import (
	"encoding/base64"
	"testing"

	"crypt-server/internal/crypto"
	"crypt-server/internal/migrate"
	"crypt-server/internal/store"
	"github.com/stretchr/testify/require"
)

func TestCreateFirstAdmin(t *testing.T) {
	dataStore := newTestSQLiteStore(t)
	err := createFirstAdmin(dataStore, "admin", "secret")
	require.NoError(t, err)

	user, err := dataStore.GetUserByUsername("admin")
	require.NoError(t, err)
	require.True(t, user.IsStaff)
	require.True(t, user.CanApprove)
	require.True(t, user.LocalLoginEnabled)
	require.False(t, user.MustResetPassword)
	require.Equal(t, "local", user.AuthSource)
}

func TestCreateFirstAdminRejectsExistingUsers(t *testing.T) {
	dataStore := newTestSQLiteStore(t)
	_, err := dataStore.AddUser("existing", "hash", true, true, true, false, "local")
	require.NoError(t, err)

	err = createFirstAdmin(dataStore, "admin", "secret")
	require.Error(t, err)
}

func TestCreateFirstAdminRequiresUsername(t *testing.T) {
	dataStore := newTestSQLiteStore(t)
	err := createFirstAdmin(dataStore, " ", "secret")
	require.Error(t, err)
}

func TestCreateFirstAdminRequiresPassword(t *testing.T) {
	dataStore := newTestSQLiteStore(t)
	err := createFirstAdmin(dataStore, "admin", "")
	require.Error(t, err)
}

func TestResetUserPassword(t *testing.T) {
	dataStore := newTestSQLiteStore(t)
	_, err := dataStore.AddUser("testuser", "oldhash", false, false, true, false, "local")
	require.NoError(t, err)

	err = resetUserPassword(dataStore, "testuser", "newpassword")
	require.NoError(t, err)

	user, err := dataStore.GetUserByUsername("testuser")
	require.NoError(t, err)
	require.NotEqual(t, "oldhash", user.PasswordHash)
}

func TestResetUserPasswordRequiresUsername(t *testing.T) {
	dataStore := newTestSQLiteStore(t)
	err := resetUserPassword(dataStore, " ", "newpassword")
	require.Error(t, err)
}

func TestResetUserPasswordRequiresPassword(t *testing.T) {
	dataStore := newTestSQLiteStore(t)
	_, err := dataStore.AddUser("testuser", "oldhash", false, false, true, false, "local")
	require.NoError(t, err)

	err = resetUserPassword(dataStore, "testuser", "")
	require.Error(t, err)
}

func TestResetUserPasswordUserNotFound(t *testing.T) {
	dataStore := newTestSQLiteStore(t)
	err := resetUserPassword(dataStore, "nonexistent", "newpassword")
	require.Error(t, err)
}

func newTestSQLiteStore(t *testing.T) *store.SQLiteStore {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	codec, err := crypto.NewAesGcmCodecFromBase64Key(encoded)
	require.NoError(t, err)
	sqliteStore, err := store.NewSQLiteStore(t.TempDir()+"/crypt.db", codec)
	require.NoError(t, err)
	migrationFS, err := migrate.SubMigrationsFS(migrate.EmbeddedFS, "sqlite")
	require.NoError(t, err)
	require.NoError(t, migrate.Apply(sqliteStore.DB(), "sqlite", migrationFS))
	return sqliteStore
}
