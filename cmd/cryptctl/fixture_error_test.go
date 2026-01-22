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

func TestConvertFixturePlaintextSecretFallback(t *testing.T) {
	// Test that non-Fernet secrets are treated as plaintext (fallback behavior)
	// This supports Django databases that didn't have field encryption enabled
	legacyKey := fernet.Key{}
	require.NoError(t, legacyKey.Generate())
	legacyDecoded := fernet.MustDecodeKeys(string(legacyKey.Encode()))

	codec, err := crypto.NewAesGcmCodecFromBase64Key(validKeyBase64())
	require.NoError(t, err)

	entries := []fixtureEntry{
		{Model: "server.secret", PK: 20, Fields: map[string]interface{}{"computer": 10, "secret_type": "recovery_key", "secret": "plaintext-recovery-key"}},
	}

	output, err := convertFixture(entries, legacyDecoded[0], codec, map[string]passwordMapEntry{})
	require.NoError(t, err)
	require.Len(t, output.Secrets, 1)
	// Verify the plaintext was encrypted with the new codec
	decrypted, err := codec.Decrypt(output.Secrets[0].Secret)
	require.NoError(t, err)
	require.Equal(t, "plaintext-recovery-key", decrypted)
}

func TestConvertFixtureMissingSecret(t *testing.T) {
	legacyKey := fernet.Key{}
	require.NoError(t, legacyKey.Generate())
	legacyDecoded := fernet.MustDecodeKeys(string(legacyKey.Encode()))

	codec, err := crypto.NewAesGcmCodecFromBase64Key(validKeyBase64())
	require.NoError(t, err)

	entries := []fixtureEntry{
		{Model: "server.secret", PK: 21, Fields: map[string]interface{}{"computer": 10, "secret_type": "recovery_key"}},
	}

	_, err = convertFixture(entries, legacyDecoded[0], codec, map[string]passwordMapEntry{})
	require.Error(t, err)
}

func TestRunImportFixtureMissingArgs(t *testing.T) {
	err := runImportFixture([]string{"--input", "", "--output", ""})
	require.Error(t, err)
}

func TestRunImportFixtureMissingKeys(t *testing.T) {
	tmp := t.TempDir()
	entries := []fixtureEntry{{Model: "server.computer", PK: 10, Fields: map[string]interface{}{"serial": "SERIAL"}}}
	payload, err := json.Marshal(entries)
	require.NoError(t, err)

	inputPath := filepath.Join(tmp, "fixture.json")
	outputPath := filepath.Join(tmp, "output.json")
	require.NoError(t, os.WriteFile(inputPath, payload, 0o600))

	err = runImportFixture([]string{"--input", inputPath, "--output", outputPath})
	require.Error(t, err)
}

func validKeyBase64() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}
