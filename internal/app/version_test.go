package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionDisplayedOnAllPages(t *testing.T) {
	server, _, sessionManager := newTestServer(t)

	// Set a test version
	originalVersion := Version
	Version = "test-version-1.2.3"
	defer func() { Version = originalVersion }()

	testCases := []struct {
		name        string
		path        string
		requireAuth bool
	}{
		{"Login page", "/login/", false},
		{"Index page", "/", true},
		{"User list", "/admin/users/", true},
		{"New user", "/admin/users/new/", true},
		{"Audit log", "/admin/audit/", true},
		{"Password change", "/password/change/", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			var req *http.Request
			if tc.requireAuth {
				req = newAuthenticatedRequest(t, sessionManager, http.MethodGet, tc.path, nil, "admin")
				serveProtected(server, rec, req, func(w http.ResponseWriter, r *http.Request) {
					// Let the actual handler run through Routes()
					server.Routes().ServeHTTP(w, r)
				})
			} else {
				req = httptest.NewRequest(http.MethodGet, tc.path, nil)
				server.Routes().ServeHTTP(rec, req)
			}

			require.Contains(t, rec.Body.String(), "Crypt Server version test-version-1.2.3",
				"Version should be displayed on %s", tc.name)
		})
	}
}

func TestVersionVariable(t *testing.T) {
	// Test that version can be set and retrieved
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "1.0.0"
	require.Equal(t, "1.0.0", Version)
}
