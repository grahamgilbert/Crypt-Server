package app

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

func hashPassword(plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("password is required")
	}
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(plaintext), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return fmt.Sprintf("$argon2id$%d$%d$%d$%s$%s",
		argon2Time,
		argon2Memory,
		argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(plaintext, encoded string) bool {
	params, salt, hash, err := parseArgon2id(encoded)
	if err != nil {
		return false
	}
	check := argon2.IDKey([]byte(plaintext), salt, params.time, params.memory, params.threads, uint32(len(hash)))
	return subtle.ConstantTimeCompare(check, hash) == 1
}

type argon2Params struct {
	time    uint32
	memory  uint32
	threads uint8
}

func parseArgon2id(encoded string) (argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 7 || parts[1] != "argon2id" {
		return argon2Params{}, nil, nil, errors.New("invalid argon2id hash")
	}
	var params argon2Params
	if _, err := fmt.Sscanf(parts[2], "%d", &params.time); err != nil {
		return argon2Params{}, nil, nil, errors.New("invalid argon2id time")
	}
	if _, err := fmt.Sscanf(parts[3], "%d", &params.memory); err != nil {
		return argon2Params{}, nil, nil, errors.New("invalid argon2id memory")
	}
	if _, err := fmt.Sscanf(parts[4], "%d", &params.threads); err != nil {
		return argon2Params{}, nil, nil, errors.New("invalid argon2id threads")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argon2Params{}, nil, nil, errors.New("invalid argon2id salt")
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[6])
	if err != nil {
		return argon2Params{}, nil, nil, errors.New("invalid argon2id hash")
	}
	return params, salt, hash, nil
}
