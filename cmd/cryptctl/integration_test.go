package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTestCheckin(t *testing.T) {
	t.Run("successful checkin", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/checkin/", r.URL.Path)
			require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

			err := r.ParseForm()
			require.NoError(t, err)
			require.Equal(t, "TEST-SERIAL", r.FormValue("serial"))
			require.Equal(t, "testuser", r.FormValue("username"))
			require.Equal(t, "Test Mac", r.FormValue("macname"))
			require.Equal(t, "secret123", r.FormValue("recovery_password"))
			require.Equal(t, "recovery_key", r.FormValue("secret_type"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"serial":            "TEST-SERIAL",
				"username":          "testuser",
				"rotation_required": false,
			})
		})

		var stdout bytes.Buffer
		err := testCheckin(handler, &stdout, "TEST-SERIAL", "testuser", "Test Mac", "secret123", "recovery_key")
		require.NoError(t, err)
		require.Contains(t, stdout.String(), "serial=TEST-SERIAL")
		require.Contains(t, stdout.String(), "username=testuser")
	})

	t.Run("checkin failure", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		})

		err := testCheckin(handler, io.Discard, "TEST-SERIAL", "testuser", "Test Mac", "secret123", "recovery_key")
		require.Error(t, err)
		require.Contains(t, err.Error(), "checkin failed with status 500")
	})

	t.Run("invalid json response", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not json"))
		})

		err := testCheckin(handler, io.Discard, "TEST-SERIAL", "testuser", "Test Mac", "secret123", "recovery_key")
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode response")
	})
}

func TestTestCheckinFormValues(t *testing.T) {
	var capturedForm map[string][]string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		capturedForm = r.Form
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"serial":            r.FormValue("serial"),
			"username":          r.FormValue("username"),
			"rotation_required": false,
		})
	})

	err := testCheckin(handler, io.Discard, "SN-12345", "jdoe", "Johns Mac", "my-secret-key", "filevault")
	require.NoError(t, err)

	require.Equal(t, []string{"SN-12345"}, capturedForm["serial"])
	require.Equal(t, []string{"jdoe"}, capturedForm["username"])
	require.Equal(t, []string{"Johns Mac"}, capturedForm["macname"])
	require.Equal(t, []string{"my-secret-key"}, capturedForm["recovery_password"])
	require.Equal(t, []string{"filevault"}, capturedForm["secret_type"])
}

func TestCheckinHTTPMethod(t *testing.T) {
	var capturedMethod string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"serial": "X", "username": "Y", "rotation_required": false})
	})

	req := httptest.NewRequest(http.MethodPost, "/checkin/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.MethodPost, capturedMethod)
}
