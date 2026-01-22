package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminUserListUI(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("testuser", passwordHash, false, true, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/users/", nil, "admin")
	serveProtected(server, rec, req, server.handleUserList)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "Users")
	require.Contains(t, body, "New User")
	require.Contains(t, body, "admin")
	require.Contains(t, body, "testuser")
	require.Contains(t, body, "/admin/users/new/")
}

func TestAdminNewUserUI(t *testing.T) {
	server, _, sessionManager := newTestServer(t)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/users/new/", nil, "admin")
	serveProtected(server, rec, req, server.handleNewUser)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "New User")
	require.Contains(t, body, "Username")
	require.Contains(t, body, "Password")
	require.Contains(t, body, "Enable local login")
	require.Contains(t, body, "Admin user")
	require.Contains(t, body, "Can approve requests")
	require.Contains(t, body, "Auth source")
}

func TestAdminEditUserUI(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	user, err := memStore.AddUser("edituser", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/users/"+intToString(user.ID)+"/edit/", nil, "admin")
	serveProtected(server, rec, req, server.handleUserEdit)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "Edit User")
	require.Contains(t, body, "edituser")
	require.Contains(t, body, "Admin user")
	require.Contains(t, body, "Can approve requests")
	require.Contains(t, body, "Local login enabled")
	require.Contains(t, body, "Save Changes")
	require.Contains(t, body, "Back")
}

func TestAdminResetPasswordUI(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	user, err := memStore.AddUser("resetuser", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/users/"+intToString(user.ID)+"/password/", nil, "admin")
	serveProtected(server, rec, req, server.handleUserPassword)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "Reset Password")
	require.Contains(t, body, "resetuser")
	require.Contains(t, body, "New password")
	require.Contains(t, body, "Reset Password")
}

func TestAdminDeleteUserUI(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	user, err := memStore.AddUser("deleteuser", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/users/"+intToString(user.ID)+"/delete/", nil, "admin")
	serveProtected(server, rec, req, server.handleUserDelete)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "Delete User")
	require.Contains(t, body, "deleteuser")
	require.Contains(t, body, "Are you sure")
	require.Contains(t, body, "Delete User")
	require.Contains(t, body, "Cancel")
}

func TestAdminAuditLogUI(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	_, err := memStore.AddAuditEvent("admin", "testuser", "user_created", "test reason", "127.0.0.1")
	require.NoError(t, err)
	_, err = memStore.AddAuditEvent("admin", "testuser2", "password_reset", "", "192.168.1.1")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/audit/", nil, "admin")
	serveProtected(server, rec, req, server.handleAuditLog)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "Audit Log")
	require.Contains(t, body, "Search audit log")
	require.Contains(t, body, "admin")
	require.Contains(t, body, "testuser")
	require.Contains(t, body, "user_created")
	require.Contains(t, body, "password_reset")
	require.Contains(t, body, "127.0.0.1")
	require.Contains(t, body, "192.168.1.1")
}

func TestAdminAuditLogSearch(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	_, err := memStore.AddAuditEvent("admin", "user1", "user_created", "", "127.0.0.1")
	require.NoError(t, err)
	_, err = memStore.AddAuditEvent("admin", "user2", "password_reset", "", "127.0.0.1")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/audit/?q=password", nil, "admin")
	serveProtected(server, rec, req, server.handleAuditLog)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "password_reset")
	require.Contains(t, body, "Clear")
}

func TestAdminAuditLogPagination(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	// Create more than 50 events to test pagination
	for i := 0; i < 55; i++ {
		_, err := memStore.AddAuditEvent("admin", "user", "user_created", "", "127.0.0.1")
		require.NoError(t, err)
	}

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/admin/audit/", nil, "admin")
	serveProtected(server, rec, req, server.handleAuditLog)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "Page 1 of 2")
	require.Contains(t, body, "Next")
}

func TestAdminUINavigation(t *testing.T) {
	server, _, sessionManager := newTestServer(t)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/", nil, "admin")
	serveProtected(server, rec, req, server.handleIndex)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "/admin/users/")
	require.Contains(t, body, "/admin/audit/")
	require.Contains(t, body, "Users")
	require.Contains(t, body, "Audit Log")
}

func TestAdminUIHiddenForNonStaffUsers(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)
	passwordHash := hashPasswordForTest(t, "password")
	_, err := memStore.AddUser("regularuser", passwordHash, false, false, true, false, "local")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := newAuthenticatedRequest(t, sessionManager, http.MethodGet, "/", nil, "regularuser")
	serveProtected(server, rec, req, server.handleIndex)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.NotContains(t, body, "/admin/users/")
	require.NotContains(t, body, "Audit Log")
}

func TestCreateUserWithAllPermissions(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)

	form := url.Values{}
	form.Set("username", "fulluser")
	form.Set("password", "Str0ng!Passw0rd")
	form.Set("local_login_enabled", "on")
	form.Set("is_staff", "on")
	form.Set("can_approve", "on")
	form.Set("auth_source", "local")

	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/new/", form, "admin")
	serveProtected(server, rec, req, server.handleNewUser)

	require.Equal(t, http.StatusSeeOther, rec.Code)

	user, err := memStore.GetUserByUsername("fulluser")
	require.NoError(t, err)
	require.True(t, user.IsStaff)
	require.True(t, user.CanApprove)
	require.True(t, user.LocalLoginEnabled)
	require.False(t, user.MustResetPassword)
	require.Equal(t, "local", user.AuthSource)
}

func TestCreateSAMLOnlyUser(t *testing.T) {
	server, memStore, sessionManager := newTestServer(t)

	form := url.Values{}
	form.Set("username", "samluser")
	form.Set("password", "")
	form.Set("is_staff", "on")
	form.Set("auth_source", "saml")

	rec := httptest.NewRecorder()
	req := newAuthenticatedFormRequest(t, server, sessionManager, http.MethodPost, "/admin/users/new/", form, "admin")
	serveProtected(server, rec, req, server.handleNewUser)

	require.Equal(t, http.StatusSeeOther, rec.Code)

	user, err := memStore.GetUserByUsername("samluser")
	require.NoError(t, err)
	require.True(t, user.IsStaff)
	require.False(t, user.LocalLoginEnabled)
	require.Equal(t, "saml", user.AuthSource)
	require.Equal(t, "", user.PasswordHash)
}
