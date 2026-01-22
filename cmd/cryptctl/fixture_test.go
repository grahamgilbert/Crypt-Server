package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"crypt-server/internal/crypto"
	"github.com/fernet/fernet-go"
	"github.com/stretchr/testify/require"
)

func TestParseFixture(t *testing.T) {
	entries := []fixtureEntry{{Model: "server.computer", PK: 1, Fields: map[string]interface{}{"serial": "ABC"}}}
	data, err := json.Marshal(entries)
	require.NoError(t, err)

	parsed, err := parseFixture(data)
	require.NoError(t, err)
	require.Len(t, parsed, 1)
	require.Equal(t, "server.computer", parsed[0].Model)
}

func TestConvertFixture(t *testing.T) {
	legacyKey := fernet.Key{}
	require.NoError(t, legacyKey.Generate())
	legacyKeyEncoded := legacyKey.Encode()
	legacyDecoded := fernet.MustDecodeKeys(string(legacyKeyEncoded))

	newCodec := testCodec(t)

	ciphertext, err := fernet.EncryptAndSign([]byte("secret"), &legacyKey)
	require.NoError(t, err)
	entries := []fixtureEntry{
		{Model: "auth.user", PK: 1, Fields: map[string]interface{}{"username": "admin", "email": "admin@example.com", "is_staff": true, "is_superuser": true}},
		{Model: "auth.group", PK: 2, Fields: map[string]interface{}{"name": "approvers"}},
		{Model: "auth.permission", PK: 3, Fields: map[string]interface{}{"codename": "can_approve"}},
		{Model: "auth.group_permissions", PK: 4, Fields: map[string]interface{}{"group": 2, "permission": 3}},
		{Model: "auth.user_groups", PK: 5, Fields: map[string]interface{}{"user": 1, "group": 2}},
		{Model: "server.computer", PK: 10, Fields: map[string]interface{}{"serial": "SERIAL", "username": "user", "computername": "Mac", "last_checkin": "2024-01-01T00:00:00Z"}},
		{Model: "server.secret", PK: 20, Fields: map[string]interface{}{"computer": 10, "secret_type": "recovery_key", "secret": string(ciphertext), "date_escrowed": "2024-01-01T00:00:00Z", "rotation_required": false}},
		{Model: "server.request", PK: 30, Fields: map[string]interface{}{"secret": 20, "requesting_user": 1, "approved": true, "auth_user": 1, "reason_for_request": "Need access", "reason_for_approval": "ok", "date_requested": "2024-01-01T00:00:00Z", "date_approved": "2024-01-01T01:00:00Z", "current": true}},
		{Model: "server.request", PK: 31, Fields: map[string]interface{}{"secret": 20, "requesting_user": 1, "approved": false, "auth_user": 1, "reason_for_request": "Need access again", "reason_for_approval": "denied", "date_requested": "2024-01-02T00:00:00Z", "date_approved": "2024-01-02T01:00:00Z", "current": false}},
	}

	passwordHash, err := hashPasswordForExport("Str0ng!Passw0rd")
	require.NoError(t, err)
	passwordMap := map[string]passwordMapEntry{
		"admin": {PasswordHash: passwordHash, MustResetPassword: false},
	}

	output, err := convertFixture(entries, legacyDecoded[0], newCodec, passwordMap)
	require.NoError(t, err)
	require.Len(t, output.Computers, 1)
	require.Len(t, output.Secrets, 1)
	require.Len(t, output.Requests, 2)
	require.Len(t, output.Users, 1)

	decrypted, err := newCodec.Decrypt(output.Secrets[0].Secret)
	require.NoError(t, err)
	require.Equal(t, "secret", decrypted)
	require.Equal(t, "admin", output.Requests[0].RequestingUser)
	require.Equal(t, "admin", output.Requests[0].AuthUser)
	require.Equal(t, false, *output.Requests[1].Approved)
	require.Equal(t, "denied", output.Requests[1].ReasonForApproval)
	require.True(t, output.Users[0].CanApprove)
	require.Equal(t, []string{"approvers"}, output.Users[0].Groups)
	require.True(t, output.Users[0].LocalLoginEnabled)
	require.Equal(t, "local", output.Users[0].AuthSource)
}

func TestRunImportFixture(t *testing.T) {
	tmp := t.TempDir()
	legacyKey := fernet.Key{}
	require.NoError(t, legacyKey.Generate())
	legacyKeyEncoded := legacyKey.Encode()

	newKey := make([]byte, 32)
	for i := range newKey {
		newKey[i] = byte(i + 1)
	}
	newKeyEncoded := base64.StdEncoding.EncodeToString(newKey)

	ciphertext, err := fernet.EncryptAndSign([]byte("secret"), &legacyKey)
	require.NoError(t, err)
	entries := []fixtureEntry{
		{Model: "server.computer", PK: 10, Fields: map[string]interface{}{"serial": "SERIAL", "username": "user", "computername": "Mac"}},
		{Model: "server.secret", PK: 20, Fields: map[string]interface{}{"computer": 10, "secret_type": "recovery_key", "secret": string(ciphertext)}},
	}
	payload, err := json.Marshal(entries)
	require.NoError(t, err)

	inputPath := filepath.Join(tmp, "fixture.json")
	outputPath := filepath.Join(tmp, "output.json")
	legacyKeyPath := filepath.Join(tmp, "legacy.key")
	newKeyPath := filepath.Join(tmp, "new.key")

	require.NoError(t, os.WriteFile(inputPath, payload, 0o600))
	require.NoError(t, os.WriteFile(legacyKeyPath, []byte(legacyKeyEncoded), 0o600))
	require.NoError(t, os.WriteFile(newKeyPath, []byte(newKeyEncoded), 0o600))

	err = runImportFixture([]string{
		"--input", inputPath,
		"--output", outputPath,
		"--legacy-key-file", legacyKeyPath,
		"--new-key-file", newKeyPath,
	})
	require.NoError(t, err)

	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	var output migrationOutput
	require.NoError(t, json.Unmarshal(outputBytes, &output))
	require.Len(t, output.Secrets, 1)

	codec, err := crypto.NewAesGcmCodecFromBase64Key(newKeyEncoded)
	require.NoError(t, err)
	plaintext, err := codec.Decrypt(output.Secrets[0].Secret)
	require.NoError(t, err)
	require.Equal(t, "secret", plaintext)
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
