package store

import (
	"bytes"
	"log"
	"path/filepath"
	"testing"
	"time"

	"crypt-server/internal/migrate"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func newTestLoggingStore(t *testing.T) (*LoggingStore, *bytes.Buffer) {
	t.Helper()
	codec := testCodec(t)
	path := filepath.Join(t.TempDir(), "crypt.db")
	sqliteStore, err := NewSQLiteStore(path, codec)
	require.NoError(t, err)

	// Apply real migrations
	sqliteFS, err := migrate.SubMigrationsFS(migrate.EmbeddedFS, "sqlite")
	require.NoError(t, err)
	err = migrate.Apply(sqliteStore.DB(), "sqlite", sqliteFS)
	require.NoError(t, err)

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	return NewLoggingStore(sqliteStore, logger), &buf
}

func TestLoggingStoreLogsWrites(t *testing.T) {
	loggingStore, buf := newTestLoggingStore(t)

	// AddComputer should log
	_, err := loggingStore.AddComputer("ABC123", "testuser", "Test-Mac")
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: AddComputer")
	require.Contains(t, buf.String(), "serial=ABC123")
	buf.Reset()

	// UpsertComputer should log
	_, err = loggingStore.UpsertComputer("DEF456", "testuser", "Test-Mac", time.Now())
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: UpsertComputer")
	buf.Reset()

	// AddUser should log
	user, err := loggingStore.AddUser("admin", "hash", true, true, true, false, "local")
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: AddUser")
	require.Contains(t, buf.String(), "username=admin")
	buf.Reset()

	// UpdateUser should log
	_, err = loggingStore.UpdateUser(user.ID, "admin", true, true, true, false, "local")
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: UpdateUser")
	buf.Reset()

	// DeleteUser should log
	err = loggingStore.DeleteUser(user.ID)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: DeleteUser")
	buf.Reset()
}

func TestLoggingStoreDoesNotLogReads(t *testing.T) {
	loggingStore, buf := newTestLoggingStore(t)

	// Add some data first
	_, _ = loggingStore.AddComputer("ABC123", "testuser", "Test-Mac")
	_, _ = loggingStore.AddUser("admin", "hash", true, true, true, false, "local")
	buf.Reset()

	// ListComputers should not log
	_, _ = loggingStore.ListComputers()
	require.Empty(t, buf.String())

	// GetComputerByID should not log
	_, _ = loggingStore.GetComputerByID(1)
	require.Empty(t, buf.String())

	// GetComputerBySerial should not log
	_, _ = loggingStore.GetComputerBySerial("ABC123")
	require.Empty(t, buf.String())

	// GetUserByUsername should not log
	_, _ = loggingStore.GetUserByUsername("admin")
	require.Empty(t, buf.String())

	// ListUsers should not log
	_, _ = loggingStore.ListUsers()
	require.Empty(t, buf.String())

	// GetUserByID should not log
	_, _ = loggingStore.GetUserByID(1)
	require.Empty(t, buf.String())
}

func TestLoggingStoreLogsSecretOperations(t *testing.T) {
	loggingStore, buf := newTestLoggingStore(t)

	// Setup
	computer, _ := loggingStore.AddComputer("ABC123", "testuser", "Test-Mac")
	buf.Reset()

	// AddSecret (new) should log
	secret, isNew, err := loggingStore.AddSecret(computer.ID, "recovery_key", "SECRET123", false)
	require.NoError(t, err)
	require.True(t, isNew)
	require.Contains(t, buf.String(), "db: AddSecret")
	require.Contains(t, buf.String(), "(new)")
	buf.Reset()

	// AddSecret with same value should log as "(updated)" (duplicate detection)
	_, isNew, err = loggingStore.AddSecret(computer.ID, "recovery_key", "SECRET123", false)
	require.NoError(t, err)
	require.False(t, isNew)
	require.Contains(t, buf.String(), "db: AddSecret")
	require.Contains(t, buf.String(), "(updated)")
	buf.Reset()

	// SetSecretRotationRequired should log
	_, err = loggingStore.SetSecretRotationRequired(secret.ID, true)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: SetSecretRotationRequired")
	buf.Reset()

	// GetSecretByID should not log
	_, _ = loggingStore.GetSecretByID(secret.ID)
	require.Empty(t, buf.String())

	// ListSecretsByComputer should not log
	_, _ = loggingStore.ListSecretsByComputer(computer.ID)
	require.Empty(t, buf.String())
}

func TestLoggingStoreLogsRequestOperations(t *testing.T) {
	loggingStore, buf := newTestLoggingStore(t)

	// Setup
	computer, _ := loggingStore.AddComputer("ABC123", "testuser", "Test-Mac")
	secret, _, _ := loggingStore.AddSecret(computer.ID, "recovery_key", "SECRET123", false)
	buf.Reset()

	// AddRequest should log
	request, err := loggingStore.AddRequest(secret.ID, "requester", "need key", "", nil)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: AddRequest")
	require.Contains(t, buf.String(), "user=requester")
	buf.Reset()

	// ApproveRequest should log
	_, err = loggingStore.ApproveRequest(request.ID, true, "approved", "approver")
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: ApproveRequest")
	require.Contains(t, buf.String(), "approved=true")
	buf.Reset()

	// GetRequestByID should not log
	_, _ = loggingStore.GetRequestByID(request.ID)
	require.Empty(t, buf.String())

	// ListRequestsBySecret should not log
	_, _ = loggingStore.ListRequestsBySecret(secret.ID)
	require.Empty(t, buf.String())

	// ListOutstandingRequests should not log
	_, _ = loggingStore.ListOutstandingRequests()
	require.Empty(t, buf.String())
}

func TestLoggingStoreLogsAuditEvents(t *testing.T) {
	loggingStore, buf := newTestLoggingStore(t)

	// AddAuditEvent should log
	_, err := loggingStore.AddAuditEvent("admin", "user1", "user_created", "", "127.0.0.1")
	require.NoError(t, err)
	require.Contains(t, buf.String(), "db: AddAuditEvent")
	require.Contains(t, buf.String(), "actor=admin")
	require.Contains(t, buf.String(), "action=user_created")
	buf.Reset()

	// ListAuditEvents should not log
	_, _ = loggingStore.ListAuditEvents()
	require.Empty(t, buf.String())

	// SearchAuditEvents should not log
	_, _ = loggingStore.SearchAuditEvents("admin")
	require.Empty(t, buf.String())
}
