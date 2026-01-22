package store

import "time"

// DateTimeFormat is the standard format for displaying dates (matches Django's Y-m-d H:i:s).
const DateTimeFormat = "2006-01-02 15:04:05"

type Computer struct {
	ID           int
	Serial       string
	Username     string
	ComputerName string
	LastCheckin  time.Time
}

// LastCheckinFormatted returns the last checkin time in display format.
func (c Computer) LastCheckinFormatted() string {
	return c.LastCheckin.Format(DateTimeFormat)
}

type Secret struct {
	ID               int
	ComputerID       int
	SecretType       string
	Secret           string
	DateEscrowed     time.Time
	RotationRequired bool
}

// SecretTypeDisplay returns the human-readable display name for the secret type.
func (s Secret) SecretTypeDisplay() string {
	switch s.SecretType {
	case "recovery_key":
		return "Recovery Key"
	case "password":
		return "Password"
	case "unlock_pin":
		return "Unlock PIN"
	default:
		return s.SecretType
	}
}

// DateEscrowedFormatted returns the escrow date in display format.
func (s Secret) DateEscrowedFormatted() string {
	return s.DateEscrowed.Format(DateTimeFormat)
}

type Request struct {
	ID                int
	SecretID          int
	RequestingUser    string
	Approved          *bool
	AuthUser          string
	ReasonForRequest  string
	ReasonForApproval string
	DateRequested     time.Time
	DateApproved      *time.Time
	Current           bool
}

// DateRequestedFormatted returns the request date in display format.
func (r Request) DateRequestedFormatted() string {
	return r.DateRequested.Format(DateTimeFormat)
}

// DateApprovedFormatted returns the approval date in display format, or empty string if not approved.
func (r Request) DateApprovedFormatted() string {
	if r.DateApproved == nil {
		return ""
	}
	return r.DateApproved.Format(DateTimeFormat)
}

type User struct {
	ID                int
	Username          string
	PasswordHash      string
	IsStaff           bool
	CanApprove        bool
	LocalLoginEnabled bool
	MustResetPassword bool
	AuthSource        string
}

type AuditEvent struct {
	ID         int
	Actor      string
	TargetUser string
	Action     string
	Reason     string
	IPAddress  string
	CreatedAt  time.Time
}

// CreatedAtFormatted returns the created time in display format.
func (a AuditEvent) CreatedAtFormatted() string {
	return a.CreatedAt.Format(DateTimeFormat)
}
