package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsUnsafeMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		{http.MethodGet, false},
		{http.MethodHead, false},
		{http.MethodOptions, false},
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodPatch, true},
		{http.MethodDelete, true},
		{"CUSTOM", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			require.Equal(t, tt.expected, isUnsafeMethod(tt.method))
		})
	}
}

func TestWithTrailingSlashRedirect(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := withTrailingSlashRedirect(handler)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedPath   string
	}{
		{"root path", "/", http.StatusOK, ""},
		{"path with trailing slash", "/foo/", http.StatusOK, ""},
		{"path without trailing slash", "/foo", http.StatusMovedPermanently, "/foo/"},
		{"nested path without slash", "/foo/bar", http.StatusMovedPermanently, "/foo/bar/"},
		{"nested path with slash", "/foo/bar/", http.StatusOK, ""},
		{"static path without slash", "/static/css", http.StatusOK, ""},
		{"static path with slash", "/static/css/", http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedPath != "" {
				require.Equal(t, tt.expectedPath, rec.Header().Get("Location"))
			}
		})
	}
}

func TestUrlQueryEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/foo/bar/", "%2Ffoo%2Fbar%2F"},
		{"hello world", "hello+world"},
		{"special=chars&more", "special%3Dchars%26more"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expected, urlQueryEscape(tt.input))
		})
	}
}

func TestServerIsCSRFExempt(t *testing.T) {
	s := &Server{}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/checkin/", true},
		{"/checkin/foo", true},
		{"/verify/", true},
		{"/verify/ABC123", true},
		{"/saml/", true},
		{"/saml/acs", true},
		{"/saml2/", true},
		{"/saml2/acs", true},
		{"/login/", false},
		{"/", false},
		{"/admin/users/", false},
		{"/request/", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			require.Equal(t, tt.expected, s.isCSRFExempt(req))
		})
	}
}
