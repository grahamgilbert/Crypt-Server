package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"crypt-server/internal/crypto"
	"crypt-server/internal/migrate"
	"crypt-server/internal/store"

	"github.com/stretchr/testify/require"
)

func TestImportFixture(t *testing.T) {
	codec, err := crypto.NewAesGcmCodecFromBase64Key("ija/CsKe9xs4RSia1SY/oVwMzMR2t5Fh3gd1GggbocY=")
	require.NoError(t, err)

	t.Run("successful import on empty database", func(t *testing.T) {
		st, cleanup := setupTestStore(t, codec)
		defer cleanup()

		// Encrypt a test secret using the same codec
		encryptedSecret, err := codec.Encrypt("test-recovery-key-12345")
		require.NoError(t, err)

		// Create a test fixture file with properly encrypted secret
		fixtureData := `{
			"computers": [
				{"id": 1, "serial": "ABC123", "username": "testuser", "computername": "Test Mac", "last_checkin": "2024-01-15T10:30:00Z"}
			],
			"secrets": [
				{"id": 1, "computer_id": 1, "secret_type": "recovery_key", "secret": "` + encryptedSecret + `", "date_escrowed": "2024-01-15T10:30:00Z", "rotation_required": false}
			],
			"users": [
				{"id": 1, "username": "admin", "email": "admin@example.com", "is_staff": true, "is_superuser": true, "can_approve": true, "groups": [], "password_hash": "", "must_reset_password": true, "local_login_enabled": false, "auth_source": "saml"}
			],
			"requests": [
				{"id": 1, "secret_id": 1, "requesting_user": "admin", "approved": true, "auth_user": "admin", "reason_for_request": "Test", "reason_for_approval": "Approved", "date_requested": "2024-01-15T11:00:00Z", "date_approved": "2024-01-15T11:05:00Z", "current": true}
			]
		}`

		fixturePath := filepath.Join(t.TempDir(), "fixture.json")
		err = os.WriteFile(fixturePath, []byte(fixtureData), 0o600)
		require.NoError(t, err)

		err = importFixture(st, fixturePath)
		require.NoError(t, err)

		// Verify computer was imported
		computers, err := st.ListComputers()
		require.NoError(t, err)
		require.Len(t, computers, 1)
		require.Equal(t, "ABC123", computers[0].Serial)
		require.Equal(t, "testuser", computers[0].Username)
		require.Equal(t, "Test Mac", computers[0].ComputerName)

		// Verify user was imported
		users, err := st.ListUsers()
		require.NoError(t, err)
		require.Len(t, users, 1)
		require.Equal(t, "admin", users[0].Username)
		require.True(t, users[0].IsStaff)
		require.True(t, users[0].CanApprove)

		// Verify secret was imported and can be decrypted
		secrets, err := st.ListSecretsByComputer(1)
		require.NoError(t, err)
		require.Len(t, secrets, 1)
		require.Equal(t, "recovery_key", secrets[0].SecretType)
		require.Equal(t, "test-recovery-key-12345", secrets[0].Secret) // Decrypted value

		// Verify request was imported
		requests, err := st.ListRequestsBySecret(1)
		require.NoError(t, err)
		require.Len(t, requests, 1)
		require.Equal(t, "admin", requests[0].RequestingUser)
		require.NotNil(t, requests[0].Approved)
		require.True(t, *requests[0].Approved)
	})

	t.Run("import fails on non-empty database", func(t *testing.T) {
		st, cleanup := setupTestStore(t, codec)
		defer cleanup()

		// Add existing data
		_, err := st.AddUser("existing", "", true, false, false, false, "local")
		require.NoError(t, err)

		fixtureData := `{"computers": [], "secrets": [], "users": [], "requests": []}`
		fixturePath := filepath.Join(t.TempDir(), "fixture.json")
		err = os.WriteFile(fixturePath, []byte(fixtureData), 0o600)
		require.NoError(t, err)

		err = importFixture(st, fixturePath)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrDatabaseNotEmpty)
	})

	t.Run("import fails with invalid JSON", func(t *testing.T) {
		st, cleanup := setupTestStore(t, codec)
		defer cleanup()

		fixturePath := filepath.Join(t.TempDir(), "fixture.json")
		err := os.WriteFile(fixturePath, []byte("not valid json"), 0o600)
		require.NoError(t, err)

		err = importFixture(st, fixturePath)
		require.Error(t, err)
		require.Contains(t, err.Error(), "parse fixture")
	})

	t.Run("import fails with missing file", func(t *testing.T) {
		st, cleanup := setupTestStore(t, codec)
		defer cleanup()

		err := importFixture(st, "/nonexistent/path/fixture.json")
		require.Error(t, err)
		require.Contains(t, err.Error(), "read fixture file")
	})
}

func TestParseDateTimeFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "RFC3339",
			input:    "2024-01-15T10:30:00Z",
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "RFC3339 with milliseconds",
			input:    "2024-01-15T10:30:00.000Z",
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "space-separated datetime",
			input:    "2024-01-15 10:30:00",
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "T-separated without Z",
			input:    "2024-01-15T10:30:00",
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Time{},
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateTime(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	codec, err := crypto.NewAesGcmCodecFromBase64Key("ija/CsKe9xs4RSia1SY/oVwMzMR2t5Fh3gd1GggbocY=")
	require.NoError(t, err)

	t.Run("returns true for empty database", func(t *testing.T) {
		st, cleanup := setupTestStore(t, codec)
		defer cleanup()

		isEmpty, err := st.IsEmpty()
		require.NoError(t, err)
		require.True(t, isEmpty)
	})

	t.Run("returns false when users exist", func(t *testing.T) {
		st, cleanup := setupTestStore(t, codec)
		defer cleanup()

		_, err := st.AddUser("testuser", "", false, false, false, false, "local")
		require.NoError(t, err)

		isEmpty, err := st.IsEmpty()
		require.NoError(t, err)
		require.False(t, isEmpty)
	})

	t.Run("returns false when computers exist", func(t *testing.T) {
		st, cleanup := setupTestStore(t, codec)
		defer cleanup()

		_, err := st.AddComputer("SERIAL123", "user", "Computer")
		require.NoError(t, err)

		isEmpty, err := st.IsEmpty()
		require.NoError(t, err)
		require.False(t, isEmpty)
	})
}

func setupTestStore(t *testing.T, codec store.SecretCodec) (store.Store, func()) {
	t.Helper()

	st, err := store.NewSQLiteStore(":memory:", codec)
	require.NoError(t, err)

	migrationsFS, err := migrate.SubMigrationsFS(migrate.EmbeddedFS, "sqlite")
	require.NoError(t, err)

	err = migrate.Apply(st.DB(), "sqlite", migrationsFS)
	require.NoError(t, err)

	return st, func() {
		st.DB().Close()
	}
}
