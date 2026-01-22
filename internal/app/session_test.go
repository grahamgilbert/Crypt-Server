package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSessionManagerRoundTrip(t *testing.T) {
	manager, err := NewSessionManager([]byte("test-session-key-32-bytes-long!!"), "crypt_session", time.Hour)
	require.NoError(t, err)

	token, err := manager.createAt("admin", time.Unix(1000, 0))
	require.NoError(t, err)

	username, ok := manager.validateAt(token, time.Unix(2000, 0))
	require.True(t, ok)
	require.Equal(t, "admin", username)
}

func TestSessionManagerExpired(t *testing.T) {
	manager, err := NewSessionManager([]byte("test-session-key-32-bytes-long!!"), "crypt_session", time.Hour)
	require.NoError(t, err)

	token, err := manager.createAt("admin", time.Unix(1000, 0))
	require.NoError(t, err)

	_, ok := manager.validateAt(token, time.Unix(1000+int64(2*time.Hour.Seconds()), 0))
	require.False(t, ok)
}

func TestSessionManagerInvalidSignature(t *testing.T) {
	manager, err := NewSessionManager([]byte("test-session-key-32-bytes-long!!"), "crypt_session", time.Hour)
	require.NoError(t, err)

	token, err := manager.createAt("admin", time.Unix(1000, 0))
	require.NoError(t, err)

	last := token[len(token)-1]
	replace := byte('a')
	if last == replace {
		replace = 'b'
	}
	tampered := token[:len(token)-1] + string(replace)
	_, ok := manager.validateAt(tampered, time.Unix(1000, 0))
	require.False(t, ok)
}
