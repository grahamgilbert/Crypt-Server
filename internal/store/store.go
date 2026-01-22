package store

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrMissingCodec = errors.New("secret codec is required")

type SecretCodec interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type Store interface {
	AddComputer(serial, username, computerName string) (*Computer, error)
	UpsertComputer(serial, username, computerName string, lastCheckin time.Time) (*Computer, error)
	ListComputers() ([]*Computer, error)
	GetComputerByID(id int) (*Computer, error)
	GetComputerBySerial(serial string) (*Computer, error)
	// AddSecret adds a new secret. Returns the secret, a bool indicating if it was newly created
	// (false if the same secret value already exists), and any error.
	AddSecret(computerID int, secretType, secret string, rotationRequired bool) (*Secret, bool, error)
	ListSecretsByComputer(computerID int) ([]*Secret, error)
	GetSecretByID(id int) (*Secret, error)
	GetLatestSecretByComputerAndType(computerID int, secretType string) (*Secret, error)
	AddRequest(secretID int, requestingUser, reason string, approvedBy string, approved *bool) (*Request, error)
	ListRequestsBySecret(secretID int) ([]*Request, error)
	ListOutstandingRequests() ([]*Request, error)
	GetRequestByID(id int) (*Request, error)
	ApproveRequest(requestID int, approved bool, reason, approver string) (*Request, error)
	AddUser(username, passwordHash string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	ListUsers() ([]*User, error)
	GetUserByID(id int) (*User, error)
	UpdateUser(id int, username string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) (*User, error)
	UpdateUserPassword(id int, passwordHash string, mustResetPassword bool) (*User, error)
	DeleteUser(id int) error
	CleanupRequests(approvedBefore time.Time) (int, error)
	SetSecretRotationRequired(secretID int, rotationRequired bool) (*Secret, error)
	AddAuditEvent(actor, targetUser, action, reason, ipAddress string) (*AuditEvent, error)
	ListAuditEvents() ([]*AuditEvent, error)
	SearchAuditEvents(query string) ([]*AuditEvent, error)
	ListAuditEventsPaged(limit, offset int) ([]*AuditEvent, error)
	SearchAuditEventsPaged(query string, limit, offset int) ([]*AuditEvent, error)
	CountAuditEvents() (int, error)
	CountSearchAuditEvents(query string) (int, error)
	// IsEmpty returns true if all data tables are empty (no rows).
	// This is used to check if it's safe to import fixture data.
	IsEmpty() (bool, error)
	// ImportComputer inserts a computer with a specific ID.
	ImportComputer(id int, serial, username, computerName string, lastCheckin time.Time) error
	// ImportSecret inserts a secret with a specific ID. The secret is already encrypted.
	ImportSecret(id, computerID int, secretType, encryptedSecret string, dateEscrowed time.Time, rotationRequired bool) error
	// ImportRequest inserts a request with a specific ID.
	ImportRequest(id, secretID int, requestingUser string, approved *bool, authUser, reasonForRequest, reasonForApproval string, dateRequested time.Time, dateApproved *time.Time, current bool) error
	// ImportUser inserts a user with a specific ID.
	ImportUser(id int, username, passwordHash string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) error
}
