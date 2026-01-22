package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type AesGcmCodec struct {
	aead cipher.AEAD
}

func NewAesGcmCodecFromBase64Key(keyBase64 string) (*AesGcmCodec, error) {
	if keyBase64 == "" {
		return nil, errors.New("missing encryption key")
	}
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}
	return NewAesGcmCodec(key)
}

func NewAesGcmCodec(key []byte) (*AesGcmCodec, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	return &AesGcmCodec{aead: aead}, nil
}

func (c *AesGcmCodec) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	ciphertext := c.aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func (c *AesGcmCodec) Decrypt(ciphertext string) (string, error) {
	payload, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	nonceSize := c.aead.NonceSize()
	if len(payload) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce := payload[:nonceSize]
	sealed := payload[nonceSize:]
	plaintext, err := c.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}
