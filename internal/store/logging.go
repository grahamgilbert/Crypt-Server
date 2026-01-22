package store

import (
	"log"
	"time"
)

// LoggingStore wraps a Store and logs all operations.
type LoggingStore struct {
	store  Store
	logger *log.Logger
}

// NewLoggingStore creates a new LoggingStore that wraps the given store.
func NewLoggingStore(store Store, logger *log.Logger) *LoggingStore {
	return &LoggingStore{store: store, logger: logger}
}

func (s *LoggingStore) AddComputer(serial, username, computerName string) (*Computer, error) {
	result, err := s.store.AddComputer(serial, username, computerName)
	if err != nil {
		s.logger.Printf("db: AddComputer failed: serial=%s error=%v", serial, err)
	} else {
		s.logger.Printf("db: AddComputer: serial=%s username=%s computername=%s", serial, username, computerName)
	}
	return result, err
}

func (s *LoggingStore) UpsertComputer(serial, username, computerName string, lastCheckin time.Time) (*Computer, error) {
	result, err := s.store.UpsertComputer(serial, username, computerName, lastCheckin)
	if err != nil {
		s.logger.Printf("db: UpsertComputer failed: serial=%s error=%v", serial, err)
	} else {
		s.logger.Printf("db: UpsertComputer: serial=%s username=%s computername=%s", serial, username, computerName)
	}
	return result, err
}

func (s *LoggingStore) ListComputers() ([]*Computer, error) {
	return s.store.ListComputers()
}

func (s *LoggingStore) GetComputerByID(id int) (*Computer, error) {
	return s.store.GetComputerByID(id)
}

func (s *LoggingStore) GetComputerBySerial(serial string) (*Computer, error) {
	return s.store.GetComputerBySerial(serial)
}

func (s *LoggingStore) AddSecret(computerID int, secretType, secret string, rotationRequired bool) (*Secret, bool, error) {
	result, isNew, err := s.store.AddSecret(computerID, secretType, secret, rotationRequired)
	if err != nil {
		s.logger.Printf("db: AddSecret failed: computer_id=%d type=%s error=%v", computerID, secretType, err)
	} else if isNew {
		s.logger.Printf("db: AddSecret: computer_id=%d type=%s (new)", computerID, secretType)
	} else {
		s.logger.Printf("db: AddSecret: computer_id=%d type=%s (updated)", computerID, secretType)
	}
	return result, isNew, err
}

func (s *LoggingStore) ListSecretsByComputer(computerID int) ([]*Secret, error) {
	return s.store.ListSecretsByComputer(computerID)
}

func (s *LoggingStore) GetSecretByID(id int) (*Secret, error) {
	return s.store.GetSecretByID(id)
}

func (s *LoggingStore) GetLatestSecretByComputerAndType(computerID int, secretType string) (*Secret, error) {
	return s.store.GetLatestSecretByComputerAndType(computerID, secretType)
}

func (s *LoggingStore) AddRequest(secretID int, requestingUser, reason string, approvedBy string, approved *bool) (*Request, error) {
	result, err := s.store.AddRequest(secretID, requestingUser, reason, approvedBy, approved)
	if err != nil {
		s.logger.Printf("db: AddRequest failed: secret_id=%d user=%s error=%v", secretID, requestingUser, err)
	} else {
		s.logger.Printf("db: AddRequest: secret_id=%d user=%s", secretID, requestingUser)
	}
	return result, err
}

func (s *LoggingStore) ListRequestsBySecret(secretID int) ([]*Request, error) {
	return s.store.ListRequestsBySecret(secretID)
}

func (s *LoggingStore) ListOutstandingRequests() ([]*Request, error) {
	return s.store.ListOutstandingRequests()
}

func (s *LoggingStore) GetRequestByID(id int) (*Request, error) {
	return s.store.GetRequestByID(id)
}

func (s *LoggingStore) ApproveRequest(requestID int, approved bool, reason, approver string) (*Request, error) {
	result, err := s.store.ApproveRequest(requestID, approved, reason, approver)
	if err != nil {
		s.logger.Printf("db: ApproveRequest failed: request_id=%d error=%v", requestID, err)
	} else {
		s.logger.Printf("db: ApproveRequest: request_id=%d approved=%t approver=%s", requestID, approved, approver)
	}
	return result, err
}

func (s *LoggingStore) AddUser(username, passwordHash string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) (*User, error) {
	result, err := s.store.AddUser(username, passwordHash, isStaff, canApprove, localLoginEnabled, mustResetPassword, authSource)
	if err != nil {
		s.logger.Printf("db: AddUser failed: username=%s error=%v", username, err)
	} else {
		s.logger.Printf("db: AddUser: username=%s is_staff=%t can_approve=%t auth_source=%s", username, isStaff, canApprove, authSource)
	}
	return result, err
}

func (s *LoggingStore) GetUserByUsername(username string) (*User, error) {
	return s.store.GetUserByUsername(username)
}

func (s *LoggingStore) ListUsers() ([]*User, error) {
	return s.store.ListUsers()
}

func (s *LoggingStore) GetUserByID(id int) (*User, error) {
	return s.store.GetUserByID(id)
}

func (s *LoggingStore) UpdateUser(id int, username string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) (*User, error) {
	result, err := s.store.UpdateUser(id, username, isStaff, canApprove, localLoginEnabled, mustResetPassword, authSource)
	if err != nil {
		s.logger.Printf("db: UpdateUser failed: id=%d username=%s error=%v", id, username, err)
	} else {
		s.logger.Printf("db: UpdateUser: id=%d username=%s is_staff=%t can_approve=%t", id, username, isStaff, canApprove)
	}
	return result, err
}

func (s *LoggingStore) UpdateUserPassword(id int, passwordHash string, mustResetPassword bool) (*User, error) {
	result, err := s.store.UpdateUserPassword(id, passwordHash, mustResetPassword)
	if err != nil {
		s.logger.Printf("db: UpdateUserPassword failed: id=%d error=%v", id, err)
	} else {
		s.logger.Printf("db: UpdateUserPassword: id=%d username=%s", id, result.Username)
	}
	return result, err
}

func (s *LoggingStore) DeleteUser(id int) error {
	err := s.store.DeleteUser(id)
	if err != nil {
		s.logger.Printf("db: DeleteUser failed: id=%d error=%v", id, err)
	} else {
		s.logger.Printf("db: DeleteUser: id=%d", id)
	}
	return err
}

func (s *LoggingStore) CleanupRequests(approvedBefore time.Time) (int, error) {
	count, err := s.store.CleanupRequests(approvedBefore)
	if err != nil {
		s.logger.Printf("db: CleanupRequests failed: error=%v", err)
	} else if count > 0 {
		s.logger.Printf("db: CleanupRequests: cleaned=%d", count)
	}
	return count, err
}

func (s *LoggingStore) SetSecretRotationRequired(secretID int, rotationRequired bool) (*Secret, error) {
	result, err := s.store.SetSecretRotationRequired(secretID, rotationRequired)
	if err != nil {
		s.logger.Printf("db: SetSecretRotationRequired failed: secret_id=%d error=%v", secretID, err)
	} else {
		s.logger.Printf("db: SetSecretRotationRequired: secret_id=%d rotation_required=%t", secretID, rotationRequired)
	}
	return result, err
}

func (s *LoggingStore) AddAuditEvent(actor, targetUser, action, reason, ipAddress string) (*AuditEvent, error) {
	result, err := s.store.AddAuditEvent(actor, targetUser, action, reason, ipAddress)
	if err != nil {
		s.logger.Printf("db: AddAuditEvent failed: actor=%s action=%s error=%v", actor, action, err)
	} else {
		s.logger.Printf("db: AddAuditEvent: actor=%s target=%s action=%s", actor, targetUser, action)
	}
	return result, err
}

func (s *LoggingStore) ListAuditEvents() ([]*AuditEvent, error) {
	return s.store.ListAuditEvents()
}

func (s *LoggingStore) SearchAuditEvents(query string) ([]*AuditEvent, error) {
	return s.store.SearchAuditEvents(query)
}

func (s *LoggingStore) ListAuditEventsPaged(limit, offset int) ([]*AuditEvent, error) {
	return s.store.ListAuditEventsPaged(limit, offset)
}

func (s *LoggingStore) SearchAuditEventsPaged(query string, limit, offset int) ([]*AuditEvent, error) {
	return s.store.SearchAuditEventsPaged(query, limit, offset)
}

func (s *LoggingStore) CountAuditEvents() (int, error) {
	return s.store.CountAuditEvents()
}

func (s *LoggingStore) CountSearchAuditEvents(query string) (int, error) {
	return s.store.CountSearchAuditEvents(query)
}

func (s *LoggingStore) IsEmpty() (bool, error) {
	return s.store.IsEmpty()
}

func (s *LoggingStore) ImportComputer(id int, serial, username, computerName string, lastCheckin time.Time) error {
	err := s.store.ImportComputer(id, serial, username, computerName, lastCheckin)
	if err != nil {
		s.logger.Printf("db: ImportComputer failed: id=%d serial=%s error=%v", id, serial, err)
	} else {
		s.logger.Printf("db: ImportComputer: id=%d serial=%s", id, serial)
	}
	return err
}

func (s *LoggingStore) ImportSecret(id, computerID int, secretType, encryptedSecret string, dateEscrowed time.Time, rotationRequired bool) error {
	err := s.store.ImportSecret(id, computerID, secretType, encryptedSecret, dateEscrowed, rotationRequired)
	if err != nil {
		s.logger.Printf("db: ImportSecret failed: id=%d computer_id=%d error=%v", id, computerID, err)
	} else {
		s.logger.Printf("db: ImportSecret: id=%d computer_id=%d type=%s", id, computerID, secretType)
	}
	return err
}

func (s *LoggingStore) ImportRequest(id, secretID int, requestingUser string, approved *bool, authUser, reasonForRequest, reasonForApproval string, dateRequested time.Time, dateApproved *time.Time, current bool) error {
	err := s.store.ImportRequest(id, secretID, requestingUser, approved, authUser, reasonForRequest, reasonForApproval, dateRequested, dateApproved, current)
	if err != nil {
		s.logger.Printf("db: ImportRequest failed: id=%d secret_id=%d error=%v", id, secretID, err)
	} else {
		s.logger.Printf("db: ImportRequest: id=%d secret_id=%d user=%s", id, secretID, requestingUser)
	}
	return err
}

func (s *LoggingStore) ImportUser(id int, username, passwordHash string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) error {
	err := s.store.ImportUser(id, username, passwordHash, isStaff, canApprove, localLoginEnabled, mustResetPassword, authSource)
	if err != nil {
		s.logger.Printf("db: ImportUser failed: id=%d username=%s error=%v", id, username, err)
	} else {
		s.logger.Printf("db: ImportUser: id=%d username=%s", id, username)
	}
	return err
}
