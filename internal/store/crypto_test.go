package store

import (
	"encoding/base64"
	"testing"

	"crypt-server/internal/crypto"
	"github.com/stretchr/testify/require"
)

func TestAesGcmCodecRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	codec, err := crypto.NewAesGcmCodecFromBase64Key(encoded)
	require.NoError(t, err)

	ciphertext, err := codec.Encrypt("secret")
	require.NoError(t, err)
	require.NotEmpty(t, ciphertext)
	require.NotEqual(t, "secret", ciphertext)

	plaintext, err := codec.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, "secret", plaintext)
}
