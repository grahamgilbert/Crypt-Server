package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCSRFBlocksMissingToken(t *testing.T) {
	server, _, _ := newTestServer(t)

	form := url.Values{}
	form.Set("field", "value")
	req := httptest.NewRequest(http.MethodPost, "/new/computer/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server.withCSRF(handler).ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFAcceptsValidToken(t *testing.T) {
	server, _, _ := newTestServer(t)
	csrfToken, err := server.csrfManager.GenerateToken()
	require.NoError(t, err)

	form := url.Values{}
	form.Set("field", "value")
	form.Set("csrf_token", csrfToken)
	req := httptest.NewRequest(http.MethodPost, "/new/computer/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: server.csrfManager.cookieName, Value: csrfToken})
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server.withCSRF(handler).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
