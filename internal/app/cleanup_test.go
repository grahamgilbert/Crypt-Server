package app

import (
	"io"
	"log"
	"testing"
	"time"

	"crypt-server/internal/store"
	"github.com/stretchr/testify/require"
)

type cleanupStoreStub struct {
	store.Store
	called bool
	cutoff time.Time
}

func (s *cleanupStoreStub) CleanupRequests(approvedBefore time.Time) (int, error) {
	s.called = true
	s.cutoff = approvedBefore
	return 0, nil
}

func TestCleanupOldRequestsUsesExpectedCutoff(t *testing.T) {
	stub := &cleanupStoreStub{}
	server := &Server{
		store:  stub,
		logger: log.New(io.Discard, "", 0),
	}

	before := time.Now()
	server.cleanupOldRequests()
	after := time.Now()

	require.True(t, stub.called)
	expectedLower := before.Add(-requestCleanupAfterApproval)
	expectedUpper := after.Add(-requestCleanupAfterApproval)
	require.True(t, stub.cutoff.After(expectedLower) || stub.cutoff.Equal(expectedLower))
	require.True(t, stub.cutoff.Before(expectedUpper) || stub.cutoff.Equal(expectedUpper))
}
