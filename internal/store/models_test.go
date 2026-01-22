package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestComputerLastCheckinFormatted(t *testing.T) {
	c := Computer{
		LastCheckin: time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC),
	}
	require.Equal(t, "2024-06-15 14:30:45", c.LastCheckinFormatted())
}

func TestSecretSecretTypeDisplay(t *testing.T) {
	tests := []struct {
		secretType string
		expected   string
	}{
		{"recovery_key", "Recovery Key"},
		{"password", "Password"},
		{"unlock_pin", "Unlock PIN"},
		{"custom_type", "custom_type"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.secretType, func(t *testing.T) {
			s := Secret{SecretType: tt.secretType}
			require.Equal(t, tt.expected, s.SecretTypeDisplay())
		})
	}
}

func TestSecretDateEscrowedFormatted(t *testing.T) {
	s := Secret{
		DateEscrowed: time.Date(2024, 1, 20, 9, 15, 30, 0, time.UTC),
	}
	require.Equal(t, "2024-01-20 09:15:30", s.DateEscrowedFormatted())
}

func TestRequestDateRequestedFormatted(t *testing.T) {
	r := Request{
		DateRequested: time.Date(2024, 3, 10, 16, 45, 0, 0, time.UTC),
	}
	require.Equal(t, "2024-03-10 16:45:00", r.DateRequestedFormatted())
}

func TestRequestDateApprovedFormatted(t *testing.T) {
	t.Run("with date", func(t *testing.T) {
		approvedTime := time.Date(2024, 3, 10, 17, 0, 0, 0, time.UTC)
		r := Request{DateApproved: &approvedTime}
		require.Equal(t, "2024-03-10 17:00:00", r.DateApprovedFormatted())
	})

	t.Run("nil date", func(t *testing.T) {
		r := Request{DateApproved: nil}
		require.Equal(t, "", r.DateApprovedFormatted())
	})
}

func TestAuditEventCreatedAtFormatted(t *testing.T) {
	a := AuditEvent{
		CreatedAt: time.Date(2024, 12, 25, 12, 0, 0, 0, time.UTC),
	}
	require.Equal(t, "2024-12-25 12:00:00", a.CreatedAtFormatted())
}

func TestDateTimeFormatConstant(t *testing.T) {
	// Verify the format constant matches Django's Y-m-d H:i:s
	require.Equal(t, "2006-01-02 15:04:05", DateTimeFormat)
}
