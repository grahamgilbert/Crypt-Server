package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStaticPathsNotRedirected(t *testing.T) {
	// Test the middleware directly to ensure static paths don't get redirected
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	wrapped := withTrailingSlashRedirect(handler)

	tests := []struct {
		name               string
		path               string
		shouldRedirect     bool
		expectedLocation   string
	}{
		{
			name:           "Static CSS not redirected",
			path:           "/static/style.css",
			shouldRedirect: false,
		},
		{
			name:           "Static nested path not redirected",
			path:           "/static/bootstrap/css/bootstrap.min.css",
			shouldRedirect: false,
		},
		{
			name:           "Static JS not redirected",
			path:           "/static/js/app.js",
			shouldRedirect: false,
		},
		{
			name:             "Non-static path redirected",
			path:             "/admin/users",
			shouldRedirect:   true,
			expectedLocation: "/admin/users/",
		},
		{
			name:             "Login path redirected",
			path:             "/login",
			shouldRedirect:   true,
			expectedLocation: "/login/",
		},
		{
			name:           "Root path not redirected",
			path:           "/",
			shouldRedirect: false,
		},
		{
			name:           "Path with trailing slash not redirected",
			path:           "/admin/users/",
			shouldRedirect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called = false
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if tt.shouldRedirect {
				require.Equal(t, http.StatusMovedPermanently, rec.Code, "Expected redirect for %s", tt.path)
				require.Equal(t, tt.expectedLocation, rec.Header().Get("Location"), "Expected redirect location")
				require.False(t, called, "Handler should not be called on redirect")
			} else {
				require.Equal(t, http.StatusOK, rec.Code, "Expected no redirect for %s", tt.path)
				require.Empty(t, rec.Header().Get("Location"), "Should not have Location header")
				require.True(t, called, "Handler should be called when not redirecting")
			}
		})
	}
}

func TestTrailingSlashRedirectWorksForNonStatic(t *testing.T) {
	server, _, _ := newTestServer(t)
	handler := server.Routes()

	// Test that non-static paths still get redirected
	req := httptest.NewRequest("GET", "/admin/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusMovedPermanently, rec.Code, "Non-static paths should redirect to add trailing slash")
	require.Equal(t, "/admin/users/", rec.Header().Get("Location"), "Should redirect to path with trailing slash")
}
