package store

import (
	"encoding/base64"
	"testing"

	"crypt-server/internal/crypto"
	"github.com/stretchr/testify/require"
)

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
