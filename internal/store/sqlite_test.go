package store

import (
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStoreAddComputer(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	codec := testCodec(t)
	store := NewSQLiteStoreWithDB(db, codec)
	lastCheckin := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO computers (serial, username, computername, last_checkin) VALUES (?, ?, ?, ?) RETURNING id, last_checkin",
	)).WithArgs("SERIAL", "user", "Mac", sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id", "last_checkin"}).AddRow(1, lastCheckin))

	computer, err := store.AddComputer("SERIAL", "user", "Mac")
	require.NoError(t, err)
	require.Equal(t, 1, computer.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewSQLiteStoreRequiresDSN(t *testing.T) {
	_, err := NewSQLiteStore("", testCodec(t))
	require.Error(t, err)
}

func TestNewSQLiteStoreOpensDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "crypt.db")

	store, err := NewSQLiteStore(path, testCodec(t))
	require.NoError(t, err)
	require.NotNil(t, store)
	require.NoError(t, store.db.Close())
}

func TestSQLiteStoreDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	require.NotNil(t, store.DB())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreUpsertComputer(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	lastCheckin := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO computers (serial, username, computername, last_checkin) VALUES (?, ?, ?, ?) ON CONFLICT(serial) DO UPDATE SET username = excluded.username, computername = excluded.computername, last_checkin = excluded.last_checkin RETURNING id, last_checkin",
	)).WithArgs("SERIAL", "user", "Mac", lastCheckin).WillReturnRows(sqlmock.NewRows([]string{"id", "last_checkin"}).AddRow(1, lastCheckin))

	computer, err := store.UpsertComputer("SERIAL", "user", "Mac", lastCheckin)
	require.NoError(t, err)
	require.Equal(t, 1, computer.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreListComputers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, serial, username, computername, last_checkin FROM computers ORDER BY id",
	)).WillReturnRows(sqlmock.NewRows([]string{"id", "serial", "username", "computername", "last_checkin"}).AddRow(1, "SERIAL", "user", "Mac", now))

	computers, err := store.ListComputers()
	require.NoError(t, err)
	require.Len(t, computers, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreGetComputerByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, serial, username, computername, last_checkin FROM computers WHERE id = ?",
	)).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "serial", "username", "computername", "last_checkin"}).AddRow(1, "SERIAL", "user", "Mac", time.Now()))

	computer, err := store.GetComputerByID(1)
	require.NoError(t, err)
	require.Equal(t, "SERIAL", computer.Serial)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreAddSecret(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	now := time.Now()
	// Expect the duplicate check query first (returns no rows = no duplicate)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required FROM secrets WHERE computer_id = ? AND secret_type = ?",
	)).WithArgs(1, "password").WillReturnRows(sqlmock.NewRows([]string{"id", "computer_id", "secret_type", "secret", "date_escrowed", "rotation_required"}))
	// Then expect the insert
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO secrets (computer_id, secret_type, secret, date_escrowed, rotation_required) VALUES (?, ?, ?, ?, ?) RETURNING id, date_escrowed",
	)).WithArgs(1, "password", sqlmock.AnyArg(), sqlmock.AnyArg(), false).WillReturnRows(sqlmock.NewRows([]string{"id", "date_escrowed"}).AddRow(5, now))

	secret, isNew, err := store.AddSecret(1, "password", "secret", false)
	require.NoError(t, err)
	require.Equal(t, 5, secret.ID)
	require.True(t, isNew)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreAddRequestAndApprove(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO requests (secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id, date_requested, date_approved",
	)).WithArgs(9, "user", sqlmock.AnyArg(), sqlmock.AnyArg(), "reason", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), true).WillReturnRows(sqlmock.NewRows([]string{"id", "date_requested", "date_approved"}).AddRow(3, now, nil))

	request, err := store.AddRequest(9, "user", "reason", "", nil)
	require.NoError(t, err)
	require.Equal(t, 3, request.ID)

	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE requests SET approved = ?, reason_for_approval = ?, auth_user = ?, date_approved = ? WHERE id = ?",
	)).WithArgs(true, "ok", "admin", sqlmock.AnyArg(), 3).WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current FROM requests WHERE id = ?",
	)).WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"id", "secret_id", "requesting_user", "approved", "auth_user", "reason_for_request", "reason_for_approval", "date_requested", "date_approved", "current"}).AddRow(3, 9, "user", true, "admin", "reason", "ok", now, now, true))

	approved, err := store.ApproveRequest(3, true, "ok", "admin")
	require.NoError(t, err)
	require.NotNil(t, approved.Approved)
	require.True(t, *approved.Approved)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreGetSecretAndRequests(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	codec := testCodec(t)
	store := NewSQLiteStoreWithDB(db, codec)
	now := time.Now()

	encryptedSecret, err := codec.Encrypt("secret")
	require.NoError(t, err)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required FROM secrets WHERE id = ?",
	)).WithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"id", "computer_id", "secret_type", "secret", "date_escrowed", "rotation_required"}).AddRow(2, 1, "password", encryptedSecret, now, 0))

	secret, err := store.GetSecretByID(2)
	require.NoError(t, err)
	require.Equal(t, "password", secret.SecretType)
	require.Equal(t, "secret", secret.Secret)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required FROM secrets WHERE computer_id = ? AND secret_type = ? ORDER BY date_escrowed DESC LIMIT 1",
	)).WithArgs(1, "password").WillReturnRows(sqlmock.NewRows([]string{"id", "computer_id", "secret_type", "secret", "date_escrowed", "rotation_required"}).AddRow(3, 1, "password", encryptedSecret, now, 1))

	latest, err := store.GetLatestSecretByComputerAndType(1, "password")
	require.NoError(t, err)
	require.Equal(t, 3, latest.ID)
	require.True(t, latest.RotationRequired)

	encryptedSecret2, err := codec.Encrypt("secret")
	require.NoError(t, err)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required FROM secrets WHERE computer_id = ? ORDER BY id",
	)).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "computer_id", "secret_type", "secret", "date_escrowed", "rotation_required"}).AddRow(2, 1, "password", encryptedSecret2, now, 0))

	secrets, err := store.ListSecretsByComputer(1)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	require.Equal(t, "secret", secrets[0].Secret)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current FROM requests WHERE secret_id = ? ORDER BY id",
	)).WithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"id", "secret_id", "requesting_user", "approved", "auth_user", "reason_for_request", "reason_for_approval", "date_requested", "date_approved", "current"}).AddRow(5, 2, "user", nil, nil, "reason", nil, now, nil, true))

	requests, err := store.ListRequestsBySecret(2)
	require.NoError(t, err)
	require.Len(t, requests, 1)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current FROM requests WHERE current = true AND approved IS NULL ORDER BY id",
	)).WillReturnRows(sqlmock.NewRows([]string{"id", "secret_id", "requesting_user", "approved", "auth_user", "reason_for_request", "reason_for_approval", "date_requested", "date_approved", "current"}).AddRow(5, 2, "user", nil, nil, "reason", nil, now, nil, true))

	outstanding, err := store.ListOutstandingRequests()
	require.NoError(t, err)
	require.Len(t, outstanding, 1)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current FROM requests WHERE id = ?",
	)).WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "secret_id", "requesting_user", "approved", "auth_user", "reason_for_request", "reason_for_approval", "date_requested", "date_approved", "current"}).AddRow(5, 2, "user", nil, nil, "reason", nil, now, nil, true))

	request, err := store.GetRequestByID(5)
	require.NoError(t, err)
	require.Equal(t, 2, request.SecretID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreUserLifecycle(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO users (username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id",
	)).WithArgs("admin", sqlmock.AnyArg(), true, true, true, false, "local").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	user, err := store.AddUser("admin", "hash", true, true, true, false, "local")
	require.NoError(t, err)
	require.Equal(t, 7, user.ID)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source FROM users WHERE lower(username) = lower(?)",
	)).WithArgs("admin").WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "is_staff", "can_approve", "local_login_enabled", "must_reset_password", "auth_source"}).AddRow(7, "admin", "hash", true, true, true, false, "local"))

	loaded, err := store.GetUserByUsername("admin")
	require.NoError(t, err)
	require.Equal(t, "admin", loaded.Username)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source FROM users WHERE id = ?",
	)).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "is_staff", "can_approve", "local_login_enabled", "must_reset_password", "auth_source"}).AddRow(7, "admin", "hash", true, true, true, false, "local"))

	byID, err := store.GetUserByID(7)
	require.NoError(t, err)
	require.Equal(t, "admin", byID.Username)

	mock.ExpectQuery(regexp.QuoteMeta(
		"UPDATE users SET username = ?, is_staff = ?, can_approve = ?, local_login_enabled = ?, must_reset_password = ?, auth_source = ? WHERE id = ? RETURNING id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source",
	)).WithArgs("updated", false, false, false, true, "local", 7).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "is_staff", "can_approve", "local_login_enabled", "must_reset_password", "auth_source"}).AddRow(7, "updated", "hash", false, false, false, true, "local"))

	updated, err := store.UpdateUser(7, "updated", false, false, false, true, "local")
	require.NoError(t, err)
	require.Equal(t, "updated", updated.Username)

	mock.ExpectQuery(regexp.QuoteMeta(
		"UPDATE users SET password_hash = ?, must_reset_password = ?, local_login_enabled = CASE WHEN ? IS NULL THEN local_login_enabled ELSE 1 END WHERE id = ? RETURNING id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source",
	)).WithArgs(sqlmock.AnyArg(), false, sqlmock.AnyArg(), 7).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "is_staff", "can_approve", "local_login_enabled", "must_reset_password", "auth_source"}).AddRow(7, "updated", "newhash", false, false, true, false, "local"))

	passwordUpdated, err := store.UpdateUserPassword(7, "newhash", false)
	require.NoError(t, err)
	require.Equal(t, "newhash", passwordUpdated.PasswordHash)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source FROM users ORDER BY id",
	)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "is_staff", "can_approve", "local_login_enabled", "must_reset_password", "auth_source"}).
		AddRow(7, "admin", "hash", true, true, true, false, "local").
		AddRow(8, "viewer", nil, false, false, false, false, "saml"))

	users, err := store.ListUsers()
	require.NoError(t, err)
	require.Len(t, users, 2)

	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM users WHERE id = ?",
	)).WithArgs(7).WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, store.DeleteUser(7))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreCleanupRequests(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE requests SET current = 0 WHERE current = 1 AND approved IS NOT NULL AND date_approved < ?",
	)).WithArgs(cutoff).WillReturnResult(sqlmock.NewResult(0, 3))

	updated, err := store.CleanupRequests(cutoff)
	require.NoError(t, err)
	require.Equal(t, 3, updated)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreSetSecretRotationRequired(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	codec := testCodec(t)
	store := NewSQLiteStoreWithDB(db, codec)
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE secrets SET rotation_required = ? WHERE id = ?",
	)).WithArgs(true, 7).WillReturnResult(sqlmock.NewResult(0, 1))

	encrypted, err := codec.Encrypt("secret")
	require.NoError(t, err)
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required FROM secrets WHERE id = ?",
	)).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "computer_id", "secret_type", "secret", "date_escrowed", "rotation_required"}).AddRow(7, 2, "password", encrypted, now, 1))

	updated, err := store.SetSecretRotationRequired(7, true)
	require.NoError(t, err)
	require.True(t, updated.RotationRequired)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreAuditEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO audit_events (actor, target_user, action, reason, ip_address) VALUES (?, ?, ?, ?, ?) RETURNING id, created_at",
	)).WithArgs("admin", "user", "password_reset", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

	event, err := store.AddAuditEvent("admin", "user", "password_reset", "", "")
	require.NoError(t, err)
	require.Equal(t, 1, event.ID)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, actor, target_user, action, reason, ip_address, created_at FROM audit_events ORDER BY created_at DESC, id DESC",
	)).WillReturnRows(sqlmock.NewRows([]string{"id", "actor", "target_user", "action", "reason", "ip_address", "created_at"}).
		AddRow(2, "admin", "user", "force_reset_enabled", nil, nil, now))

	events, err := store.ListAuditEvents()
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "force_reset_enabled", events[0].Action)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreSearchAuditEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, actor, target_user, action, reason, ip_address, created_at FROM audit_events WHERE lower(actor) LIKE lower(?) OR lower(target_user) LIKE lower(?) OR lower(action) LIKE lower(?) OR lower(COALESCE(reason, '')) LIKE lower(?) OR lower(COALESCE(ip_address, '')) LIKE lower(?) ORDER BY created_at DESC, id DESC",
	)).WithArgs("%reset%", "%reset%", "%reset%", "%reset%", "%reset%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "actor", "target_user", "action", "reason", "ip_address", "created_at"}).
			AddRow(3, "admin", "user", "password_reset", "reason", "127.0.0.1", now))

	events, err := store.SearchAuditEvents("reset")
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "password_reset", events[0].Action)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreAuditEventsPagingAndCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM audit_events",
	)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(120))

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, actor, target_user, action, reason, ip_address, created_at FROM audit_events ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?",
	)).WithArgs(50, 50).
		WillReturnRows(sqlmock.NewRows([]string{"id", "actor", "target_user", "action", "reason", "ip_address", "created_at"}).
			AddRow(10, "admin", "user", "password_reset", nil, nil, now))

	count, err := store.CountAuditEvents()
	require.NoError(t, err)
	require.Equal(t, 120, count)

	events, err := store.ListAuditEventsPaged(50, 50)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteStoreCountSearchAuditEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLiteStoreWithDB(db, testCodec(t))
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM audit_events WHERE lower(actor) LIKE lower(?) OR lower(target_user) LIKE lower(?) OR lower(action) LIKE lower(?) OR lower(COALESCE(reason, '')) LIKE lower(?) OR lower(COALESCE(ip_address, '')) LIKE lower(?)",
	)).WithArgs("%reset%", "%reset%", "%reset%", "%reset%", "%reset%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := store.CountSearchAuditEvents("reset")
	require.NoError(t, err)
	require.Equal(t, 5, count)
	require.NoError(t, mock.ExpectationsWereMet())
}
