package app

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"crypt-server/internal/crypto"
	"crypt-server/internal/migrate"
	"crypt-server/internal/store"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) (*Server, store.Store, *SessionManager) {
	t.Helper()
	codec := testCodec(t)
	dataStore := newTestSQLiteStore(t, codec)
	server, sessionManager := newTestServerWithStore(t, dataStore)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := dataStore.AddUser("admin", passwordHash, true, true, true, false, "local")
	require.NoError(t, err)
	return server, dataStore, sessionManager
}

func newTestSQLiteStore(t *testing.T, codec *crypto.AesGcmCodec) *store.SQLiteStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "crypt.db")
	sqliteStore, err := store.NewSQLiteStore(path, codec)
	require.NoError(t, err)
	migrationFS, err := migrate.SubMigrationsFS(migrate.EmbeddedFS, "sqlite")
	require.NoError(t, err)
	require.NoError(t, migrate.Apply(sqliteStore.DB(), "sqlite", migrationFS))
	return sqliteStore
}

func newTestServerWithStore(t *testing.T, dataStore store.Store) (*Server, *SessionManager) {
	t.Helper()
	root := filepath.Join("..", "..")
	layout := filepath.Join(root, "web", "templates", "layouts", "base.html")
	pages := filepath.Join(root, "web", "templates", "pages")
	renderer := NewRenderer(layout, pages)
	logger := log.New(io.Discard, "", 0)
	sessionManager, err := NewSessionManager([]byte("test-session-key-32-bytes-long!!"), "crypt_session", time.Hour)
	require.NoError(t, err)
	settings := Settings{
		ApproveOwn:             true,
		AllApprove:             false,
		SessionTTL:             time.Hour,
		CookieSecure:           false,
		RequestCleanupInterval: 0,
		RotateViewedSecrets:    false,
	}
	csrfManager := NewCSRFManager("crypt_csrf", 32)
	server := NewServer(dataStore, renderer, logger, sessionManager, csrfManager, nil, nil, settings)
	return server, sessionManager
}

type rotationTrackingStore struct {
	*store.SQLiteStore
	called bool
}

func (s *rotationTrackingStore) SetSecretRotationRequired(secretID int, rotationRequired bool) (*store.Secret, error) {
	s.called = true
	return s.SQLiteStore.SetSecretRotationRequired(secretID, rotationRequired)
}

type auditPaginationStore struct {
	lastLimit  int
	lastOffset int
}

func (s *auditPaginationStore) ListAuditEventsPaged(limit, offset int) ([]*store.AuditEvent, error) {
	s.lastLimit = limit
	s.lastOffset = offset
	return []*store.AuditEvent{}, nil
}

func (s *auditPaginationStore) SearchAuditEventsPaged(query string, limit, offset int) ([]*store.AuditEvent, error) {
	s.lastLimit = limit
	s.lastOffset = offset
	return []*store.AuditEvent{}, nil
}

func (s *auditPaginationStore) CountAuditEvents() (int, error) {
	return 1, nil
}

func (s *auditPaginationStore) CountSearchAuditEvents(query string) (int, error) {
	return 1, nil
}

func (s *auditPaginationStore) ListAuditEvents() ([]*store.AuditEvent, error) {
	return []*store.AuditEvent{}, nil
}

func (s *auditPaginationStore) SearchAuditEvents(query string) ([]*store.AuditEvent, error) {
	return []*store.AuditEvent{}, nil
}

func (s *auditPaginationStore) AddAuditEvent(actor, targetUser, action, reason, ipAddress string) (*store.AuditEvent, error) {
	return &store.AuditEvent{}, nil
}

func (s *auditPaginationStore) AddComputer(serial, username, computerName string) (*store.Computer, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) UpsertComputer(serial, username, computerName string, lastCheckin time.Time) (*store.Computer, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) ListComputers() ([]*store.Computer, error) {
	return []*store.Computer{}, nil
}

func (s *auditPaginationStore) GetComputerByID(id int) (*store.Computer, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) GetComputerBySerial(serial string) (*store.Computer, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) AddSecret(computerID int, secretType, secret string, rotationRequired bool) (*store.Secret, bool, error) {
	return nil, false, store.ErrNotFound
}

func (s *auditPaginationStore) ListSecretsByComputer(computerID int) ([]*store.Secret, error) {
	return []*store.Secret{}, nil
}

func (s *auditPaginationStore) GetSecretByID(id int) (*store.Secret, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) GetLatestSecretByComputerAndType(computerID int, secretType string) (*store.Secret, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) AddRequest(secretID int, requestingUser, reason string, approvedBy string, approved *bool) (*store.Request, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) ListRequestsBySecret(secretID int) ([]*store.Request, error) {
	return []*store.Request{}, nil
}

func (s *auditPaginationStore) ListOutstandingRequests() ([]*store.Request, error) {
	return []*store.Request{}, nil
}

func (s *auditPaginationStore) GetRequestByID(id int) (*store.Request, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) ApproveRequest(requestID int, approved bool, reason, approver string) (*store.Request, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) AddUser(username, passwordHash string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) (*store.User, error) {
	return nil, nil
}

func (s *auditPaginationStore) GetUserByUsername(username string) (*store.User, error) {
	return &store.User{ID: 1, Username: username, IsStaff: true, LocalLoginEnabled: true}, nil
}

func (s *auditPaginationStore) ListUsers() ([]*store.User, error) {
	return []*store.User{}, nil
}

func (s *auditPaginationStore) GetUserByID(id int) (*store.User, error) {
	return &store.User{ID: id, Username: "user"}, nil
}

func (s *auditPaginationStore) UpdateUser(id int, username string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) (*store.User, error) {
	return &store.User{ID: id, Username: username}, nil
}

func (s *auditPaginationStore) UpdateUserPassword(id int, passwordHash string, mustResetPassword bool) (*store.User, error) {
	return &store.User{ID: id, Username: "user"}, nil
}

func (s *auditPaginationStore) DeleteUser(id int) error {
	return nil
}

func (s *auditPaginationStore) CleanupRequests(approvedBefore time.Time) (int, error) {
	return 0, nil
}

func (s *auditPaginationStore) SetSecretRotationRequired(secretID int, rotationRequired bool) (*store.Secret, error) {
	return nil, store.ErrNotFound
}

func (s *auditPaginationStore) IsEmpty() (bool, error) {
	return true, nil
}

func (s *auditPaginationStore) ImportComputer(id int, serial, username, computerName string, lastCheckin time.Time) error {
	return nil
}

func (s *auditPaginationStore) ImportSecret(id, computerID int, secretType, encryptedSecret string, dateEscrowed time.Time, rotationRequired bool) error {
	return nil
}

func (s *auditPaginationStore) ImportRequest(id, secretID int, requestingUser string, approved *bool, authUser, reasonForRequest, reasonForApproval string, dateRequested time.Time, dateApproved *time.Time, current bool) error {
	return nil
}

func (s *auditPaginationStore) ImportUser(id int, username, passwordHash string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) error {
	return nil
}

func TestHandleIndex(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	_, err := memStore.AddComputer("SERIAL1", "user", "Mac")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/", nil, "admin")
	serveProtected(server, rec, req, server.handleIndex)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Serial Number")
}

func TestHandleTableAjax(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	_, err := memStore.AddComputer("SERIAL2", "user", "iMac")
	require.NoError(t, err)

	payload := map[string]any{"draw": 1}
	payloadBytes, _ := json.Marshal(payload)
	query := url.Values{}
	query.Set("args", string(payloadBytes))

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/ajax/?"+query.Encode(), nil, "admin")
	serveProtected(server, rec, req, server.handleTableAjax)

	require.Equal(t, http.StatusOK, rec.Code)

	var data map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &data))
	require.Equal(t, float64(1), data["recordsTotal"])
}

func TestHandleNewComputerFlow(t *testing.T) {
	server, _, sessionManager := newTestServer(t)
	form := url.Values{}
	form.Set("serial", "SERIAL3")
	form.Set("username", "user3")
	form.Set("computername", "MacBook Air")

	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/new/computer/", form, "admin")
	serveProtected(server, rec, req, server.handleNewComputer)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Contains(t, rec.Header().Get("Location"), "/info/")
}

func TestRequestApproveRetrieveFlow(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL4", "user4", "MacBook Pro")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret-value", false)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("reason_for_request", "Need access")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/request/"+intToString(secret.ID)+"/", form, "admin")
	serveProtected(server, rec, req, server.handleRequest)

	require.Equal(t, http.StatusSeeOther, rec.Code)

	requests, err := memStore.ListRequestsBySecret(secret.ID)
	require.NoError(t, err)
	require.Len(t, requests, 1)

	infoRec := httptest.NewRecorder()
	infoReq := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/info/secret/"+intToString(secret.ID)+"/", nil, "admin")
	serveProtected(server, infoRec, infoReq, server.handleSecretInfo)
	require.Contains(t, infoRec.Body.String(), "Retrieve Key")

	retrieveRec := httptest.NewRecorder()
	retrieveReq := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/retrieve/"+intToString(requests[0].ID)+"/", nil, "admin")
	serveProtected(server, retrieveRec, retrieveReq, server.handleRetrieve)
	require.Equal(t, http.StatusOK, retrieveRec.Code)
	require.Contains(t, retrieveRec.Body.String(), "class=\"letter\">s")
}

func TestRetrieveMarksRotationRequired(t *testing.T) {
	codec := testCodec(t)
	sqliteStore := newTestSQLiteStore(t, codec)
	dataStore := &rotationTrackingStore{SQLiteStore: sqliteStore}
	server, sessionManager := newTestServerWithStore(t, dataStore)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := dataStore.AddUser("admin", passwordHash, true, true, true, false, "local")
	require.NoError(t, err)
	server.settings.RotateViewedSecrets = true
	require.True(t, server.settings.RotateViewedSecrets)
	computer, err := dataStore.AddComputer("SERIALROTATE", "user", "MacBook Pro")
	require.NoError(t, err)
	secret, _, err := dataStore.AddSecret(computer.ID, "recovery_key", "secret-value", false)
	require.NoError(t, err)
	initial, err := dataStore.GetSecretByID(secret.ID)
	require.NoError(t, err)
	require.False(t, initial.RotationRequired)
	approved := true
	req, err := dataStore.AddRequest(secret.ID, "admin", "Need access", "approver", &approved)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	retrieveReq := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/retrieve/"+intToString(req.ID)+"/", nil, "admin")
	serveProtected(server, rec, retrieveReq, server.handleRetrieve)
	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, dataStore.called)

	var rotation int
	row := sqliteStore.DB().QueryRow("SELECT rotation_required FROM secrets WHERE id = ?", secret.ID)
	require.NoError(t, row.Scan(&rotation))
	require.Equal(t, 1, rotation)

	updated, err := dataStore.GetSecretByID(secret.ID)
	require.NoError(t, err)
	require.True(t, updated.RotationRequired)
}

func TestHandleManageRequests(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL5", "user5", "Mac Mini")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "password", "secret", false)
	require.NoError(t, err)
	_, err = memStore.AddRequest(secret.ID, "user5", "Need access", "", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/manage-requests/", nil, "admin")
	serveProtected(server, rec, req, server.handleManageRequests)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "SERIAL5")
}

func TestUserPasswordResetLogsAuditEvent(t *testing.T) {
	codec := testCodec(t)
	sqliteStore := newTestSQLiteStore(t, codec)
	server, sessionManager := newTestServerWithStore(t, sqliteStore)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := sqliteStore.AddUser("admin", passwordHash, true, true, true, false, "local")
	require.NoError(t, err)
	target, err := sqliteStore.AddUser("reset", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("password", "newpassword")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/"+intToString(target.ID)+"/password/", form, "admin")
	req.RemoteAddr = "192.0.2.9:1234"
	serveProtected(server, rec, req, server.handleUserPassword)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	events, err := sqliteStore.ListAuditEvents()
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "admin", events[0].Actor)
	require.Equal(t, "reset", events[0].TargetUser)
	require.Equal(t, "password_reset", events[0].Action)
	require.Equal(t, "192.0.2.9", events[0].IPAddress)
}

func TestUserEditForceResetLogsAuditEvent(t *testing.T) {
	codec := testCodec(t)
	sqliteStore := newTestSQLiteStore(t, codec)
	server, sessionManager := newTestServerWithStore(t, sqliteStore)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := sqliteStore.AddUser("admin", passwordHash, true, true, true, false, "local")
	require.NoError(t, err)
	target, err := sqliteStore.AddUser("target", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("username", "target")
	form.Set("must_reset_password", "on")
	form.Set("local_login_enabled", "on")
	form.Set("auth_source", "local")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/"+intToString(target.ID)+"/edit/", form, "admin")
	req.RemoteAddr = "198.51.100.10:9999"
	serveProtected(server, rec, req, server.handleUserEdit)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	events, err := sqliteStore.ListAuditEvents()
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "force_reset_enabled", events[0].Action)
	require.Equal(t, "198.51.100.10", events[0].IPAddress)
}

func TestIDFromPath(t *testing.T) {
	id, err := idFromPath("/info/", "/info/123/")
	require.NoError(t, err)
	require.Equal(t, 123, id)

	_, err = idFromPath("/info/", "/other/123/")
	require.Error(t, err)
}

func TestAuditLogSearch(t *testing.T) {
	codec := testCodec(t)
	sqliteStore := newTestSQLiteStore(t, codec)
	server, sessionManager := newTestServerWithStore(t, sqliteStore)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := sqliteStore.AddUser("admin", passwordHash, true, true, true, false, "local")
	require.NoError(t, err)
	_, err = sqliteStore.AddAuditEvent("admin", "user1", "password_reset", "reason", "127.0.0.1")
	require.NoError(t, err)
	_, err = sqliteStore.AddAuditEvent("admin", "user2", "user_deleted", "", "127.0.0.1")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/audit/?q=reset", nil, "admin")
	serveProtected(server, rec, req, server.handleAuditLog)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "password_reset")
	require.NotContains(t, rec.Body.String(), "user_deleted")
}

func TestAuditLogPaginationUsesPageParam(t *testing.T) {
	storeSpy := &auditPaginationStore{}
	server, sessionManager := newTestServerWithStore(t, storeSpy)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/audit/?page=2", nil, "admin")
	serveProtected(server, rec, req, server.handleAuditLog)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 50, storeSpy.lastLimit)
	require.Equal(t, 50, storeSpy.lastOffset)
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.5:4444"
	require.Equal(t, "203.0.113.5", clientIP(req))

	req.Header.Set("X-Real-IP", "192.0.2.1")
	require.Equal(t, "192.0.2.1", clientIP(req))

	req.Header.Set("X-Forwarded-For", "198.51.100.1, 203.0.113.9")
	require.Equal(t, "198.51.100.1", clientIP(req))
}

func TestLookupComputer(t *testing.T) {
	server, memStore, _ := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL6", "user", "Mac Studio")
	require.NoError(t, err)

	byID, err := server.lookupComputer(intToString(computer.ID))
	require.NoError(t, err)
	require.Equal(t, "SERIAL6", byID.Serial)

	bySerial, err := server.lookupComputer("serial6")
	require.NoError(t, err)
	require.Equal(t, computer.ID, bySerial.ID)
}

func TestCheckinCreatesSecret(t *testing.T) {
	server, memStore, _ := newTestServer(t)

	form := url.Values{}
	form.Set("serial", "SERIALCHECKIN")
	form.Set("recovery_password", "secret-value")
	form.Set("username", "user1")
	form.Set("macname", "MacBook")
	form.Set("secret_type", "recovery_key")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleCheckin(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "\"serial\":\"SERIALCHECKIN\"")
	require.Contains(t, rec.Body.String(), "\"rotation_required\":false")
	require.Contains(t, rec.Body.String(), "\"new_secret_escrowed\":true")

	computer, err := memStore.GetComputerBySerial("SERIALCHECKIN")
	require.NoError(t, err)
	require.Equal(t, "user1", computer.Username)
}

func TestVerifyEscrowed(t *testing.T) {
	server, memStore, _ := newTestServer(t)
	computer, err := memStore.AddComputer("SERIALVERIFY", "user", "Mac")
	require.NoError(t, err)
	_, _, err = memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/verify/SERIALVERIFY/recovery_key/", nil)
	server.handleVerify(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "\"escrowed\":true")
	require.Contains(t, rec.Body.String(), "\"date_escrowed\"")
}

func TestVerifyNotEscrowed(t *testing.T) {
	server, memStore, _ := newTestServer(t)
	_, err := memStore.AddComputer("SERIALVERIFY2", "user", "Mac")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/verify/SERIALVERIFY2/recovery_key/", nil)
	server.handleVerify(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "\"escrowed\":false")
}

func TestVerifyMissingComputer(t *testing.T) {
	server, _, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/verify/UNKNOWN/recovery_key/", nil)
	server.handleVerify(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleLoginSuccess(t *testing.T) {
	server, _, _ := newTestServer(t)

	form := url.Values{}
	form.Set("username", "admin")
	form.Set("password", "password")
	form.Set("next", "/")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleLogin(rec, req)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.NotEmpty(t, rec.Header().Get("Set-Cookie"))
}

func TestHandleLoginFailure(t *testing.T) {
	server, _, _ := newTestServer(t)

	form := url.Values{}
	form.Set("username", "admin")
	form.Set("password", "wrong")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleLogin(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid username or password.")
}

func TestHandleLoginRequiresLocalLogin(t *testing.T) {
	server, memStore, _ := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("samluser", passwordHash, false, false, false, false, "saml")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("username", "samluser")
	form.Set("password", "password")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleLogin(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid username or password.")
}

func TestHandleLoginRedirectsToReset(t *testing.T) {
	server, memStore, _ := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("resetme", passwordHash, false, false, true, true, "local")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("username", "resetme")
	form.Set("password", "password")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleLogin(rec, req)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/password/reset/", rec.Header().Get("Location"))
}

func TestHandlePasswordChange(t *testing.T) {
	server, _, sessionManager := newTestServer(t)

	form := url.Values{}
	form.Set("current_password", "password")
	form.Set("new_password", "Str0ng!Passw0rd")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/password/change/", form, "admin")
	serveProtected(server, rec, req, server.handlePasswordChange)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/", rec.Header().Get("Location"))
}

func TestHandlePasswordResetClearsFlag(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	resetUser, err := memStore.AddUser("resetuser", passwordHash, false, false, true, true, "local")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("new_password", "Str0ng!Passw0rd")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/password/reset/", form, "resetuser")
	serveProtected(server, rec, req, server.handlePasswordReset)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	updated, err := memStore.GetUserByID(resetUser.ID)
	require.NoError(t, err)
	require.False(t, updated.MustResetPassword)
	require.True(t, verifyPassword("Str0ng!Passw0rd", updated.PasswordHash))
}

func TestHandleSAMLLoginStub(t *testing.T) {
	server, _, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/saml/login/", nil)
	server.handleSAMLLogin(rec, req)
	require.Equal(t, http.StatusNotImplemented, rec.Code)
}

func TestHandleUserListRequiresStaff(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("viewer", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/users/", nil, "viewer")
	serveProtected(server, rec, req, server.handleUserList)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleUserList(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("second", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/users/", nil, "admin")
	serveProtected(server, rec, req, server.handleUserList)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "admin")
	require.Contains(t, rec.Body.String(), "second")
}

func TestHandleNewUser(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)

	form := url.Values{}
	form.Set("username", "newuser")
	form.Set("password", "Str0ng!Passw0rd")
	form.Set("local_login_enabled", "on")
	form.Set("auth_source", "local")
	form.Set("is_staff", "on")
	form.Set("can_approve", "on")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/new/", form, "admin")
	serveProtected(server, rec, req, server.handleNewUser)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	user, err := memStore.GetUserByUsername("newuser")
	require.NoError(t, err)
	require.True(t, user.IsStaff)
	require.True(t, user.CanApprove)
	require.True(t, user.LocalLoginEnabled)
	require.Equal(t, "local", user.AuthSource)
	require.True(t, verifyPassword("Str0ng!Passw0rd", user.PasswordHash))
	events, err := memStore.ListAuditEvents()
	require.NoError(t, err)
	event, ok := findAuditEvent(events, "user_created")
	require.True(t, ok)
	require.Equal(t, "newuser", event.TargetUser)
}

func TestHandleUserEdit(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	target, err := memStore.AddUser("editor", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("username", "updated")
	form.Set("is_staff", "on")
	form.Set("can_approve", "on")
	form.Set("local_login_enabled", "on")
	form.Set("must_reset_password", "on")
	form.Set("auth_source", "saml")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/"+intToString(target.ID)+"/edit/", form, "admin")
	serveProtected(server, rec, req, server.handleUserEdit)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	updated, err := memStore.GetUserByID(target.ID)
	require.NoError(t, err)
	require.Equal(t, "updated", updated.Username)
	require.True(t, updated.IsStaff)
	require.True(t, updated.CanApprove)
	require.True(t, updated.MustResetPassword)
	require.Equal(t, "saml", updated.AuthSource)
	events, err := memStore.ListAuditEvents()
	require.NoError(t, err)
	_, ok := findAuditEvent(events, "user_updated")
	require.True(t, ok)
	_, ok = findAuditEvent(events, "force_reset_enabled")
	require.True(t, ok)
}

func TestHandleUserPassword(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	target, err := memStore.AddUser("reset", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("password", "Str0ng!Passw0rd")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/"+intToString(target.ID)+"/password/", form, "admin")
	serveProtected(server, rec, req, server.handleUserPassword)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	updated, err := memStore.GetUserByID(target.ID)
	require.NoError(t, err)
	require.True(t, verifyPassword("Str0ng!Passw0rd", updated.PasswordHash))
	require.False(t, updated.MustResetPassword)
}

func TestHandleUserDelete(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	target, err := memStore.AddUser("remove", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/"+intToString(target.ID)+"/delete/", url.Values{}, "admin")
	serveProtected(server, rec, req, server.handleUserDelete)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	_, err = memStore.GetUserByID(target.ID)
	require.Error(t, err)
	events, err := memStore.ListAuditEvents()
	require.NoError(t, err)
	event, ok := findAuditEvent(events, "user_deleted")
	require.True(t, ok)
	require.Equal(t, "remove", event.TargetUser)
}

func TestHandleUserDeleteSelf(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	admin, err := memStore.GetUserByUsername("admin")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/"+intToString(admin.ID)+"/delete/", url.Values{}, "admin")
	serveProtected(server, rec, req, server.handleUserDelete)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "You cannot delete your own account.")
}

func intToString(value int) string {
	return strconv.Itoa(value)
}

func findAuditEvent(events []*store.AuditEvent, action string) (*store.AuditEvent, bool) {
	for _, event := range events {
		if event.Action == action {
			return event, true
		}
	}
	return nil, false
}

func newAuthenticatedRequest(t *testing.T, sessionManager *SessionManager, method, target string, body io.Reader, username string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	token, err := sessionManager.Create(username)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: sessionManager.CookieName(), Value: token})
	return req
}

func newAuthenticatedFormRequest(t *testing.T, server *Server, sessionManager *SessionManager, method, target string, form url.Values, username string) *http.Request {
	t.Helper()
	csrfToken, err := server.csrfManager.GenerateToken()
	require.NoError(t, err)
	form.Set("csrf_token", csrfToken)
	req := newAuthenticatedRequest(t, sessionManager, method, target, strings.NewReader(form.Encode()), username)
	req.AddCookie(&http.Cookie{Name: server.csrfManager.cookieName, Value: csrfToken})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func serveProtected(server *Server, rec *httptest.ResponseRecorder, req *http.Request, handler http.HandlerFunc) {
	server.withCSRF(server.withUser(http.HandlerFunc(server.requireAuth(handler)))).ServeHTTP(rec, req)
}

func hashPasswordForTest(t *testing.T, password string) string {
	t.Helper()
	hash, err := hashPassword(password)
	require.NoError(t, err)
	return hash
}

func testCodec(t *testing.T) *crypto.AesGcmCodec {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	codec, err := crypto.NewAesGcmCodecFromBase64Key(encoded)
	require.NoError(t, err)
	return codec
}

// Additional comprehensive tests for missing coverage

func TestHandleNewComputerGET(t *testing.T) {
	server, _, sessionManager := newTestServer(t)
	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/new/computer/", nil, "admin")
	serveProtected(server, rec, req, server.handleNewComputer)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "New Computer")
	require.Contains(t, rec.Body.String(), "Serial")
}

func TestHandleNewComputerPOSTValidation(t *testing.T) {
	server, _, sessionManager := newTestServer(t)
	form := url.Values{}
	form.Set("serial", "")
	form.Set("computername", "")

	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/new/computer/", form, "admin")
	serveProtected(server, rec, req, server.handleNewComputer)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Serial number and computer name are required.")
}

func TestHandleNewSecretGET(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL_SECRET_GET", "user", "MacBook Pro")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/new/secret/"+intToString(computer.ID)+"/", nil, "admin")
	serveProtected(server, rec, req, server.handleNewSecret)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "New Secret")
	require.Contains(t, rec.Body.String(), "MacBook Pro")
}

func TestHandleNewSecretPOST(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL_SECRET_POST", "user", "Mac")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("secret_type", "recovery_key")
	form.Set("secret", "test-recovery-key-123456")
	form.Set("rotation_required", "on")

	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/new/secret/"+intToString(computer.ID)+"/", form, "admin")
	serveProtected(server, rec, req, server.handleNewSecret)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Contains(t, rec.Header().Get("Location"), "/info/"+intToString(computer.ID))

	secrets, err := memStore.ListSecretsByComputer(computer.ID)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	require.Equal(t, "recovery_key", secrets[0].SecretType)
	require.True(t, secrets[0].RotationRequired)
}

func TestHandleNewSecretPOSTValidation(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL_SECRET_VALID", "user", "Mac")
	require.NoError(t, err)

	form := url.Values{}
	form.Set("secret_type", "")
	form.Set("secret", "")

	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/new/secret/"+intToString(computer.ID)+"/", form, "admin")
	serveProtected(server, rec, req, server.handleNewSecret)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Secret type and value are required.")
}

func TestHandleNewSecretInvalidComputer(t *testing.T) {
	server, _, sessionManager := newTestServer(t)
	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/new/secret/99999/", nil, "admin")
	serveProtected(server, rec, req, server.handleNewSecret)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleComputerInfo(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL_INFO", "testuser", "MacBook Air")
	require.NoError(t, err)
	_, _, err = memStore.AddSecret(computer.ID, "recovery_key", "test-secret", false)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/info/"+intToString(computer.ID)+"/", nil, "admin")
	serveProtected(server, rec, req, server.handleComputerInfo)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "SERIAL_INFO")
	require.Contains(t, rec.Body.String(), "MacBook Air")
	require.Contains(t, rec.Body.String(), "testuser")
	require.Contains(t, rec.Body.String(), "Recovery Key")
}

func TestHandleComputerInfoBySerial(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	_, err := memStore.AddComputer("SERIAL_BY_SERIAL", "user", "iMac")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/info/serial_by_serial/", nil, "admin")
	serveProtected(server, rec, req, server.handleComputerInfo)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "SERIAL_BY_SERIAL")
}

func TestHandleComputerInfoNotFound(t *testing.T) {
	server, _, sessionManager := newTestServer(t)
	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/info/99999/", nil, "admin")
	serveProtected(server, rec, req, server.handleComputerInfo)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleSecretInfo(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL_SECRET_INFO", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "password", "secret-pass", false)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/info/secret/"+intToString(secret.ID)+"/", nil, "admin")
	serveProtected(server, rec, req, server.handleSecretInfo)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Secret Info")
	require.Contains(t, rec.Body.String(), "SERIAL_SECRET_INFO")
	require.Contains(t, rec.Body.String(), "Get Key")
}

func TestHandleSecretInfoWithPendingRequestNonApprover(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("viewer", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	computer, err := memStore.AddComputer("SERIAL_PENDING", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	_, err = memStore.AddRequest(secret.ID, "viewer", "Need access", "", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/info/secret/"+intToString(secret.ID)+"/", nil, "viewer")
	serveProtected(server, rec, req, server.handleSecretInfo)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotContains(t, rec.Body.String(), "Request Key")
	require.Contains(t, rec.Body.String(), "Request Pending")
}

func TestHandleRequestGET(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL_REQUEST_GET", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/request/"+intToString(secret.ID)+"/", nil, "admin")
	serveProtected(server, rec, req, server.handleRequest)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Request Secret")
	require.Contains(t, rec.Body.String(), "SERIAL_REQUEST_GET")
}

func TestHandleRequestPOSTWithoutAutoApprove(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	server.settings.ApproveOwn = false
	computer, err := memStore.AddComputer("SERIAL_NO_AUTO", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("reason_for_request", "Testing no auto-approve")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/request/"+intToString(secret.ID)+"/", form, "admin")
	serveProtected(server, rec, req, server.handleRequest)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	requests, err := memStore.ListRequestsBySecret(secret.ID)
	require.NoError(t, err)
	require.Len(t, requests, 1)
	require.Nil(t, requests[0].Approved)
}

func TestHandleApproveGET(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	server.settings.ApproveOwn = false
	computer, err := memStore.AddComputer("SERIAL_APPROVE_GET", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	request, err := memStore.AddRequest(secret.ID, "requester", "Need access", "", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/approve/"+intToString(request.ID)+"/", nil, "admin")
	serveProtected(server, rec, req, server.handleApprove)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Approve Request")
	require.Contains(t, rec.Body.String(), "SERIAL_APPROVE_GET")
	require.Contains(t, rec.Body.String(), "reason_for_approval")
}

func TestHandleApprovePOSTApprove(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	server.settings.ApproveOwn = false
	computer, err := memStore.AddComputer("SERIAL_APPROVE", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	request, err := memStore.AddRequest(secret.ID, "requester", "Need access", "", nil)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("approved", "1")
	form.Set("reason_for_approval", "Looks good")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/approve/"+intToString(request.ID)+"/", form, "admin")
	serveProtected(server, rec, req, server.handleApprove)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/manage-requests/", rec.Header().Get("Location"))

	updated, err := memStore.GetRequestByID(request.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.Approved)
	require.True(t, *updated.Approved)
	require.Equal(t, "admin", updated.AuthUser)
	require.Equal(t, "Looks good", updated.ReasonForApproval)
}

func TestHandleApprovePOSTDeny(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	server.settings.ApproveOwn = false
	computer, err := memStore.AddComputer("SERIAL_DENY", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	request, err := memStore.AddRequest(secret.ID, "requester", "Need access", "", nil)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("approved", "0")
	form.Set("reason_for_approval", "Insufficient justification")
	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/approve/"+intToString(request.ID)+"/", form, "admin")
	serveProtected(server, rec, req, server.handleApprove)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	updated, err := memStore.GetRequestByID(request.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.Approved)
	require.False(t, *updated.Approved)
	require.Equal(t, "Insufficient justification", updated.ReasonForApproval)
}

func TestHandleApproveWithoutPermission(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("noapprove", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	computer, err := memStore.AddComputer("SERIAL_NO_PERM", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	request, err := memStore.AddRequest(secret.ID, "admin", "Need access", "", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/approve/"+intToString(request.ID)+"/", nil, "noapprove")
	serveProtected(server, rec, req, server.handleApprove)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleApproveSelfRequestWhenNotAllowed(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	server.settings.ApproveOwn = false

	computer, err := memStore.AddComputer("SERIAL_SELF", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	request, err := memStore.AddRequest(secret.ID, "admin", "Need access", "", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/approve/"+intToString(request.ID)+"/", nil, "admin")
	serveProtected(server, rec, req, server.handleApprove)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleLogout(t *testing.T) {
	server, _, sessionManager := newTestServer(t)
	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/logout/", nil, "admin")
	server.handleLogout(rec, req)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/login/", rec.Header().Get("Location"))

	cookies := rec.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == sessionManager.CookieName() {
			found = true
			require.Equal(t, "", cookie.Value)
			require.Equal(t, -1, cookie.MaxAge)
		}
	}
	require.True(t, found)
}

func TestHandleRetrieveUnapprovedRequest(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL_UNAPPROVED", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	request, err := memStore.AddRequest(secret.ID, "admin", "Need access", "", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/retrieve/"+intToString(request.ID)+"/", nil, "admin")
	serveProtected(server, rec, req, server.handleRetrieve)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleRetrieveDeniedRequest(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	computer, err := memStore.AddComputer("SERIAL_DENIED", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	approved := false
	request, err := memStore.AddRequest(secret.ID, "admin", "Need access", "approver", &approved)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/retrieve/"+intToString(request.ID)+"/", nil, "admin")
	serveProtected(server, rec, req, server.handleRetrieve)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleRetrieveWrongUser(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("other", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	computer, err := memStore.AddComputer("SERIAL_WRONG_USER", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	approved := true
	request, err := memStore.AddRequest(secret.ID, "admin", "Need access", "approver", &approved)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/retrieve/"+intToString(request.ID)+"/", nil, "other")
	serveProtected(server, rec, req, server.handleRetrieve)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleManageRequestsRequiresApprover(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("viewer", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/manage-requests/", nil, "viewer")
	serveProtected(server, rec, req, server.handleManageRequests)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireAuthRedirectsToLogin(t *testing.T) {
	server, _, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	server.withUser(http.HandlerFunc(server.requireAuth(server.handleIndex))).ServeHTTP(rec, req)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Contains(t, rec.Header().Get("Location"), "/login/")
	require.Contains(t, rec.Header().Get("Location"), "next=%2F")
}

func TestCSRFProtectionBlocksInvalidToken(t *testing.T) {
	server, _, sessionManager := newTestServer(t)
	form := url.Values{}
	form.Set("serial", "TEST")
	form.Set("username", "user")
	form.Set("computername", "Mac")
	form.Set("csrf_token", "invalid-token")

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodPost, "/new/computer/", strings.NewReader(form.Encode()), "admin")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	server.withCSRF(server.withUser(http.HandlerFunc(server.requireAuth(server.handleNewComputer)))).ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "CSRF token missing or invalid")
}

func TestCheckinDefaultsSecretType(t *testing.T) {
	server, memStore, _ := newTestServer(t)

	form := url.Values{}
	form.Set("serial", "SERIAL_DEFAULT_TYPE")
	form.Set("recovery_password", "secret-value")
	form.Set("username", "user1")
	form.Set("macname", "MacBook")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleCheckin(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	computer, err := memStore.GetComputerBySerial("SERIAL_DEFAULT_TYPE")
	require.NoError(t, err)
	secret, err := memStore.GetLatestSecretByComputerAndType(computer.ID, "recovery_key")
	require.NoError(t, err)
	require.NotNil(t, secret)
}

func TestCheckinMissingRequiredFields(t *testing.T) {
	server, _, _ := newTestServer(t)

	form := url.Values{}
	form.Set("serial", "")
	form.Set("recovery_password", "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleCheckin(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestVerifyInvalidPath(t *testing.T) {
	server, _, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/verify/invalid/", nil)
	server.handleVerify(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAllApproveSettingGrantsApprovalPermission(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	server.settings.AllApprove = true
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("regularuser", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	computer, err := memStore.AddComputer("SERIAL_ALL_APPROVE", "user", "Mac")
	require.NoError(t, err)
	secret, _, err := memStore.AddSecret(computer.ID, "recovery_key", "secret", false)
	require.NoError(t, err)
	request, err := memStore.AddRequest(secret.ID, "admin", "Need access", "", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/approve/"+intToString(request.ID)+"/", nil, "regularuser")
	serveProtected(server, rec, req, server.handleApprove)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestCheckinNewSecretEscrowedFirstTime(t *testing.T) {
	server, memStore, _ := newTestServer(t)

	form := url.Values{}
	form.Set("serial", "SERIAL_NEW_SECRET")
	form.Set("recovery_password", "first-secret-value")
	form.Set("username", "user1")
	form.Set("macname", "MacBook")
	form.Set("secret_type", "recovery_key")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleCheckin(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var data map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &data))
	require.Equal(t, "SERIAL_NEW_SECRET", data["serial"])
	require.Equal(t, true, data["new_secret_escrowed"])

	computer, err := memStore.GetComputerBySerial("SERIAL_NEW_SECRET")
	require.NoError(t, err)
	secrets, err := memStore.ListSecretsByComputer(computer.ID)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
}

func TestCheckinNewSecretEscrowedDuplicate(t *testing.T) {
	server, memStore, _ := newTestServer(t)

	// First escrow
	form := url.Values{}
	form.Set("serial", "SERIAL_DUPLICATE")
	form.Set("recovery_password", "same-secret-value")
	form.Set("username", "user1")
	form.Set("macname", "MacBook")
	form.Set("secret_type", "recovery_key")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleCheckin(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var firstData map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &firstData))
	require.Equal(t, true, firstData["new_secret_escrowed"])

	// Second escrow with same secret
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleCheckin(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)

	var secondData map[string]any
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &secondData))
	require.Equal(t, false, secondData["new_secret_escrowed"])

	// Verify only one secret was created
	computer, err := memStore.GetComputerBySerial("SERIAL_DUPLICATE")
	require.NoError(t, err)
	secrets, err := memStore.ListSecretsByComputer(computer.ID)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
}

func TestCheckinNewSecretEscrowedDifferentKey(t *testing.T) {
	server, memStore, _ := newTestServer(t)

	// First escrow
	form := url.Values{}
	form.Set("serial", "SERIAL_DIFF_KEY")
	form.Set("recovery_password", "first-secret")
	form.Set("username", "user1")
	form.Set("macname", "MacBook")
	form.Set("secret_type", "recovery_key")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleCheckin(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var firstData map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &firstData))
	require.Equal(t, true, firstData["new_secret_escrowed"])

	// Second escrow with different secret value
	form.Set("recovery_password", "second-different-secret")
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleCheckin(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)

	var secondData map[string]any
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &secondData))
	require.Equal(t, true, secondData["new_secret_escrowed"])

	// Verify two secrets were created
	computer, err := memStore.GetComputerBySerial("SERIAL_DIFF_KEY")
	require.NoError(t, err)
	secrets, err := memStore.ListSecretsByComputer(computer.ID)
	require.NoError(t, err)
	require.Len(t, secrets, 2)
}
