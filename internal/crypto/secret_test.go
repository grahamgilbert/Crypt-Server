package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAesGcmCodecFromBase64Key(t *testing.T) {
	// Valid 32-byte key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	encoded := base64.StdEncoding.EncodeToString(key)

	codec, err := NewAesGcmCodecFromBase64Key(encoded)
	require.NoError(t, err)
	require.NotNil(t, codec)
}

func TestNewAesGcmCodecFromBase64KeyEmpty(t *testing.T) {
	_, err := NewAesGcmCodecFromBase64Key("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing encryption key")
}

func TestNewAesGcmCodecFromBase64KeyInvalidBase64(t *testing.T) {
	_, err := NewAesGcmCodecFromBase64Key("not-valid-base64!!!")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode key")
}

func TestNewAesGcmCodecFromBase64KeyWrongLength(t *testing.T) {
	// 16-byte key (too short)
	key := make([]byte, 16)
	encoded := base64.StdEncoding.EncodeToString(key)

	_, err := NewAesGcmCodecFromBase64Key(encoded)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid key length")
}

func TestNewAesGcmCodec(t *testing.T) {
	key := make([]byte, 32)
	codec, err := NewAesGcmCodec(key)
	require.NoError(t, err)
	require.NotNil(t, codec)
}

func TestNewAesGcmCodecWrongLength(t *testing.T) {
	tests := []struct {
		name    string
		keyLen  int
	}{
		{"too short", 16},
		{"too long", 64},
		{"empty", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLen)
			_, err := NewAesGcmCodec(key)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid key length")
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	codec, err := NewAesGcmCodec(key)
	require.NoError(t, err)

	tests := []string{
		"hello world",
		"",
		"a",
		"recovery-key-12345-ABCDE",
		"unicode: 日本語 emoji: 🔐",
		string(make([]byte, 1000)), // long string
	}

	for _, plaintext := range tests {
		t.Run(plaintext[:min(len(plaintext), 20)], func(t *testing.T) {
			ciphertext, err := codec.Encrypt(plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, ciphertext)
			require.NotEqual(t, plaintext, ciphertext)

			decrypted, err := codec.Decrypt(ciphertext)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)
		})
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	key := make([]byte, 32)
	codec, err := NewAesGcmCodec(key)
	require.NoError(t, err)

	plaintext := "same plaintext"
	ciphertext1, err := codec.Encrypt(plaintext)
	require.NoError(t, err)

	ciphertext2, err := codec.Encrypt(plaintext)
	require.NoError(t, err)

	// Due to random nonce, same plaintext should produce different ciphertext
	require.NotEqual(t, ciphertext1, ciphertext2)

	// But both should decrypt to the same value
	decrypted1, err := codec.Decrypt(ciphertext1)
	require.NoError(t, err)
	decrypted2, err := codec.Decrypt(ciphertext2)
	require.NoError(t, err)
	require.Equal(t, decrypted1, decrypted2)
}

func TestDecryptInvalidBase64(t *testing.T) {
	key := make([]byte, 32)
	codec, err := NewAesGcmCodec(key)
	require.NoError(t, err)

	_, err = codec.Decrypt("not-valid-base64!!!")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode ciphertext")
}

func TestDecryptTooShort(t *testing.T) {
	key := make([]byte, 32)
	codec, err := NewAesGcmCodec(key)
	require.NoError(t, err)

	// Base64 encode a very short payload (less than nonce size)
	short := base64.StdEncoding.EncodeToString([]byte("abc"))
	_, err = codec.Decrypt(short)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ciphertext too short")
}

func TestDecryptTampered(t *testing.T) {
	key := make([]byte, 32)
	codec, err := NewAesGcmCodec(key)
	require.NoError(t, err)

	ciphertext, err := codec.Encrypt("secret data")
	require.NoError(t, err)

	// Decode, tamper, re-encode
	payload, err := base64.StdEncoding.DecodeString(ciphertext)
	require.NoError(t, err)
	payload[len(payload)-1] ^= 0xFF // flip bits in last byte
	tampered := base64.StdEncoding.EncodeToString(payload)

	_, err = codec.Decrypt(tampered)
	require.Error(t, err)
	require.Contains(t, err.Error(), "decrypt")
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1 // different key

	codec1, err := NewAesGcmCodec(key1)
	require.NoError(t, err)
	codec2, err := NewAesGcmCodec(key2)
	require.NoError(t, err)

	ciphertext, err := codec1.Encrypt("secret")
	require.NoError(t, err)

	_, err = codec2.Decrypt(ciphertext)
	require.Error(t, err)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
