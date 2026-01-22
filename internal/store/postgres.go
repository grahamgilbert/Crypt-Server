package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db    *sql.DB
	codec SecretCodec
}

func (s *PostgresStore) DB() *sql.DB {
	return s.db
}

func NewPostgresStore(dbURL string, codec SecretCodec) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &PostgresStore{db: db, codec: codec}, nil
}

func NewPostgresStoreWithDB(db *sql.DB, codec SecretCodec) *PostgresStore {
	return &PostgresStore{db: db, codec: codec}
}

func (s *PostgresStore) AddComputer(serial, username, computerName string) (*Computer, error) {
	var id int
	var lastCheckin time.Time
	err := s.db.QueryRow(
		`INSERT INTO computers (serial, username, computername, last_checkin)
		 VALUES ($1, $2, $3, NOW())
		 RETURNING id, last_checkin`,
		serial, username, computerName,
	).Scan(&id, &lastCheckin)
	if err != nil {
		return nil, fmt.Errorf("insert computer: %w", err)
	}
	return &Computer{
		ID:           id,
		Serial:       serial,
		Username:     username,
		ComputerName: computerName,
		LastCheckin:  lastCheckin,
	}, nil
}

func (s *PostgresStore) UpsertComputer(serial, username, computerName string, lastCheckin time.Time) (*Computer, error) {
	var id int
	var stored time.Time
	err := s.db.QueryRow(
		`INSERT INTO computers (serial, username, computername, last_checkin)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (serial)
		 DO UPDATE SET username = EXCLUDED.username, computername = EXCLUDED.computername, last_checkin = EXCLUDED.last_checkin
		 RETURNING id, last_checkin`,
		serial, username, computerName, lastCheckin,
	).Scan(&id, &stored)
	if err != nil {
		return nil, fmt.Errorf("upsert computer: %w", err)
	}
	return &Computer{
		ID:           id,
		Serial:       serial,
		Username:     username,
		ComputerName: computerName,
		LastCheckin:  stored,
	}, nil
}

func (s *PostgresStore) ListComputers() ([]*Computer, error) {
	rows, err := s.db.Query(`SELECT id, serial, username, computername, last_checkin FROM computers ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list computers: %w", err)
	}
	defer rows.Close()

	computers := make([]*Computer, 0)
	for rows.Next() {
		var computer Computer
		var lastCheckin sql.NullTime
		if err := rows.Scan(&computer.ID, &computer.Serial, &computer.Username, &computer.ComputerName, &lastCheckin); err != nil {
			return nil, fmt.Errorf("scan computer: %w", err)
		}
		if lastCheckin.Valid {
			computer.LastCheckin = lastCheckin.Time
		}
		computers = append(computers, &computer)
	}
	return computers, rows.Err()
}

func (s *PostgresStore) GetComputerByID(id int) (*Computer, error) {
	var computer Computer
	var lastCheckin sql.NullTime
	row := s.db.QueryRow(`SELECT id, serial, username, computername, last_checkin FROM computers WHERE id = $1`, id)
	if err := row.Scan(&computer.ID, &computer.Serial, &computer.Username, &computer.ComputerName, &lastCheckin); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get computer by id: %w", err)
	}
	if lastCheckin.Valid {
		computer.LastCheckin = lastCheckin.Time
	}
	return &computer, nil
}

func (s *PostgresStore) GetComputerBySerial(serial string) (*Computer, error) {
	var computer Computer
	var lastCheckin sql.NullTime
	row := s.db.QueryRow(`SELECT id, serial, username, computername, last_checkin FROM computers WHERE lower(serial) = lower($1)`, serial)
	if err := row.Scan(&computer.ID, &computer.Serial, &computer.Username, &computer.ComputerName, &lastCheckin); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get computer by serial: %w", err)
	}
	if lastCheckin.Valid {
		computer.LastCheckin = lastCheckin.Time
	}
	return &computer, nil
}

func (s *PostgresStore) AddSecret(computerID int, secretType, secret string, rotationRequired bool) (*Secret, bool, error) {
	if s.codec == nil {
		return nil, false, ErrMissingCodec
	}

	// Check for duplicate secret (matching Django behavior)
	// If the same secret value already exists for this computer/type and rotation is not required, skip insert
	if !rotationRequired {
		existing, err := s.findExistingSecret(computerID, secretType, secret)
		if err != nil {
			return nil, false, err
		}
		if existing != nil {
			return existing, false, nil
		}
	}

	encrypted, err := s.codec.Encrypt(secret)
	if err != nil {
		return nil, false, err
	}
	var id int
	var dateEscrowed time.Time
	row := s.db.QueryRow(
		`INSERT INTO secrets (computer_id, secret_type, secret, date_escrowed, rotation_required)
		 VALUES ($1, $2, $3, NOW(), $4)
		 RETURNING id, date_escrowed`,
		computerID, secretType, encrypted, rotationRequired,
	)
	if err := row.Scan(&id, &dateEscrowed); err != nil {
		return nil, false, fmt.Errorf("insert secret: %w", err)
	}

	decrypted, err := s.decryptSecret(&Secret{
		ID:               id,
		ComputerID:       computerID,
		SecretType:       secretType,
		Secret:           encrypted,
		DateEscrowed:     dateEscrowed,
		RotationRequired: rotationRequired,
	})
	if err != nil {
		return nil, false, err
	}
	return decrypted, true, nil
}

// findExistingSecret checks if the same secret value already exists for this computer/type
func (s *PostgresStore) findExistingSecret(computerID int, secretType, plainSecret string) (*Secret, error) {
	rows, err := s.db.Query(
		`SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required FROM secrets WHERE computer_id = $1 AND secret_type = $2`,
		computerID, secretType,
	)
	if err != nil {
		return nil, fmt.Errorf("find existing secret: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var secret Secret
		if err := rows.Scan(&secret.ID, &secret.ComputerID, &secret.SecretType, &secret.Secret, &secret.DateEscrowed, &secret.RotationRequired); err != nil {
			return nil, fmt.Errorf("scan secret: %w", err)
		}
		decrypted, err := s.decryptSecret(&secret)
		if err != nil {
			return nil, err
		}
		if decrypted.Secret == plainSecret {
			return decrypted, nil
		}
	}
	return nil, nil
}

func (s *PostgresStore) ListSecretsByComputer(computerID int) ([]*Secret, error) {
	rows, err := s.db.Query(`SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required FROM secrets WHERE computer_id = $1 ORDER BY id`, computerID)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	defer rows.Close()

	secrets := make([]*Secret, 0)
	for rows.Next() {
		var secret Secret
		if err := rows.Scan(&secret.ID, &secret.ComputerID, &secret.SecretType, &secret.Secret, &secret.DateEscrowed, &secret.RotationRequired); err != nil {
			return nil, fmt.Errorf("scan secret: %w", err)
		}
		decrypted, err := s.decryptSecret(&secret)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, decrypted)
	}
	return secrets, rows.Err()
}

func (s *PostgresStore) GetSecretByID(id int) (*Secret, error) {
	var secret Secret
	row := s.db.QueryRow(`SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required FROM secrets WHERE id = $1`, id)
	if err := row.Scan(&secret.ID, &secret.ComputerID, &secret.SecretType, &secret.Secret, &secret.DateEscrowed, &secret.RotationRequired); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}
	return s.decryptSecret(&secret)
}

func (s *PostgresStore) GetLatestSecretByComputerAndType(computerID int, secretType string) (*Secret, error) {
	var secret Secret
	row := s.db.QueryRow(
		`SELECT id, computer_id, secret_type, secret, date_escrowed, rotation_required
		 FROM secrets
		 WHERE computer_id = $1 AND secret_type = $2
		 ORDER BY date_escrowed DESC
		 LIMIT 1`,
		computerID, secretType,
	)
	if err := row.Scan(&secret.ID, &secret.ComputerID, &secret.SecretType, &secret.Secret, &secret.DateEscrowed, &secret.RotationRequired); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get latest secret: %w", err)
	}
	return s.decryptSecret(&secret)
}

func (s *PostgresStore) AddRequest(secretID int, requestingUser, reason string, approvedBy string, approved *bool) (*Request, error) {
	var id int
	var dateRequested time.Time
	var dateApproved sql.NullTime
	var approvedValue sql.NullBool
	if approved != nil {
		approvedValue.Valid = true
		approvedValue.Bool = *approved
	}
	row := s.db.QueryRow(
		`INSERT INTO requests (secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), CASE WHEN $3 IS NULL THEN NULL ELSE NOW() END, true)
		 RETURNING id, date_requested, date_approved`,
		secretID, requestingUser, approvedValue, nullableString(approvedBy), reason, sql.NullString{},
	)
	if err := row.Scan(&id, &dateRequested, &dateApproved); err != nil {
		return nil, fmt.Errorf("insert request: %w", err)
	}

	var approvedPtr *bool
	if approvedValue.Valid {
		approvedPtr = &approvedValue.Bool
	}
	var dateApprovedPtr *time.Time
	if dateApproved.Valid {
		dateApprovedPtr = &dateApproved.Time
	}

	return &Request{
		ID:                id,
		SecretID:          secretID,
		RequestingUser:    requestingUser,
		Approved:          approvedPtr,
		AuthUser:          approvedBy,
		ReasonForRequest:  reason,
		ReasonForApproval: "",
		DateRequested:     dateRequested,
		DateApproved:      dateApprovedPtr,
		Current:           true,
	}, nil
}

func (s *PostgresStore) ListRequestsBySecret(secretID int) ([]*Request, error) {
	rows, err := s.db.Query(`SELECT id, secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current FROM requests WHERE secret_id = $1 ORDER BY id`, secretID)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	defer rows.Close()

	requests := make([]*Request, 0)
	for rows.Next() {
		request, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (s *PostgresStore) ListOutstandingRequests() ([]*Request, error) {
	rows, err := s.db.Query(`SELECT id, secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current FROM requests WHERE current = true AND approved IS NULL ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list outstanding requests: %w", err)
	}
	defer rows.Close()

	requests := make([]*Request, 0)
	for rows.Next() {
		request, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (s *PostgresStore) GetRequestByID(id int) (*Request, error) {
	row := s.db.QueryRow(`SELECT id, secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current FROM requests WHERE id = $1`, id)
	request, err := scanRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return request, nil
}

func (s *PostgresStore) ApproveRequest(requestID int, approved bool, reason, approver string) (*Request, error) {
	var dateApproved time.Time
	row := s.db.QueryRow(
		`UPDATE requests
		 SET approved = $1, reason_for_approval = $2, auth_user = $3, date_approved = NOW()
		 WHERE id = $4
		 RETURNING date_approved`,
		approved, reason, approver, requestID,
	)
	if err := row.Scan(&dateApproved); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("approve request: %w", err)
	}
	updated, err := s.GetRequestByID(requestID)
	if err != nil {
		return nil, err
	}
	updated.DateApproved = &dateApproved
	return updated, nil
}

func (s *PostgresStore) AddUser(username, passwordHash string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) (*User, error) {
	var id int
	err := s.db.QueryRow(
		`INSERT INTO users (username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		username, nullableString(passwordHash), isStaff, canApprove, localLoginEnabled, mustResetPassword, authSource,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &User{
		ID:                id,
		Username:          username,
		PasswordHash:      passwordHash,
		IsStaff:           isStaff,
		CanApprove:        canApprove,
		LocalLoginEnabled: localLoginEnabled,
		MustResetPassword: mustResetPassword,
		AuthSource:        authSource,
	}, nil
}

func (s *PostgresStore) GetUserByUsername(username string) (*User, error) {
	var user User
	var passwordHash sql.NullString
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source
		 FROM users WHERE lower(username) = lower($1)`,
		username,
	)
	if err := row.Scan(&user.ID, &user.Username, &passwordHash, &user.IsStaff, &user.CanApprove, &user.LocalLoginEnabled, &user.MustResetPassword, &user.AuthSource); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	return &user, nil
}

func (s *PostgresStore) ListUsers() ([]*User, error) {
	rows, err := s.db.Query(`SELECT id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		var user User
		var passwordHash sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &passwordHash, &user.IsStaff, &user.CanApprove, &user.LocalLoginEnabled, &user.MustResetPassword, &user.AuthSource); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		if passwordHash.Valid {
			user.PasswordHash = passwordHash.String
		}
		users = append(users, &user)
	}
	return users, rows.Err()
}

func (s *PostgresStore) GetUserByID(id int) (*User, error) {
	var user User
	var passwordHash sql.NullString
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source
		 FROM users WHERE id = $1`,
		id,
	)
	if err := row.Scan(&user.ID, &user.Username, &passwordHash, &user.IsStaff, &user.CanApprove, &user.LocalLoginEnabled, &user.MustResetPassword, &user.AuthSource); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	return &user, nil
}

func (s *PostgresStore) UpdateUser(id int, username string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) (*User, error) {
	var user User
	var passwordHash sql.NullString
	row := s.db.QueryRow(
		`UPDATE users
		 SET username = $1, is_staff = $2, can_approve = $3, local_login_enabled = $4, must_reset_password = $5, auth_source = $6
		 WHERE id = $7
		 RETURNING id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source`,
		username, isStaff, canApprove, localLoginEnabled, mustResetPassword, authSource, id,
	)
	if err := row.Scan(&user.ID, &user.Username, &passwordHash, &user.IsStaff, &user.CanApprove, &user.LocalLoginEnabled, &user.MustResetPassword, &user.AuthSource); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	return &user, nil
}

func (s *PostgresStore) UpdateUserPassword(id int, passwordHash string, mustResetPassword bool) (*User, error) {
	var user User
	var hash sql.NullString
	row := s.db.QueryRow(
		`UPDATE users
		 SET password_hash = $1, must_reset_password = $2, local_login_enabled = CASE WHEN $1 IS NULL THEN local_login_enabled ELSE true END
		 WHERE id = $3
		 RETURNING id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source`,
		nullableString(passwordHash), mustResetPassword, id,
	)
	if err := row.Scan(&user.ID, &user.Username, &hash, &user.IsStaff, &user.CanApprove, &user.LocalLoginEnabled, &user.MustResetPassword, &user.AuthSource); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update user password: %w", err)
	}
	if hash.Valid {
		user.PasswordHash = hash.String
	}
	return &user, nil
}

func (s *PostgresStore) DeleteUser(id int) error {
	result, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) CleanupRequests(approvedBefore time.Time) (int, error) {
	result, err := s.db.Exec(
		`UPDATE requests SET current = false WHERE current = true AND approved IS NOT NULL AND date_approved < $1`,
		approvedBefore,
	)
	if err != nil {
		return 0, fmt.Errorf("cleanup requests: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("cleanup requests: %w", err)
	}
	return int(affected), nil
}

func (s *PostgresStore) SetSecretRotationRequired(secretID int, rotationRequired bool) (*Secret, error) {
	var secret Secret
	row := s.db.QueryRow(
		`UPDATE secrets
		 SET rotation_required = $1
		 WHERE id = $2
		 RETURNING id, computer_id, secret_type, secret, date_escrowed, rotation_required`,
		rotationRequired, secretID,
	)
	if err := row.Scan(&secret.ID, &secret.ComputerID, &secret.SecretType, &secret.Secret, &secret.DateEscrowed, &secret.RotationRequired); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update secret rotation: %w", err)
	}
	return s.decryptSecret(&secret)
}

func (s *PostgresStore) AddAuditEvent(actor, targetUser, action, reason, ipAddress string) (*AuditEvent, error) {
	var id int
	var createdAt time.Time
	err := s.db.QueryRow(
		`INSERT INTO audit_events (actor, target_user, action, reason, ip_address)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		actor, targetUser, action, nullableString(reason), nullableString(ipAddress),
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("insert audit event: %w", err)
	}
	return &AuditEvent{
		ID:         id,
		Actor:      actor,
		TargetUser: targetUser,
		Action:     action,
		Reason:     reason,
		IPAddress:  ipAddress,
		CreatedAt:  createdAt,
	}, nil
}

func (s *PostgresStore) ListAuditEvents() ([]*AuditEvent, error) {
	rows, err := s.db.Query(`SELECT id, actor, target_user, action, reason, ip_address, created_at FROM audit_events ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	return scanAuditEvents(rows)
}

func (s *PostgresStore) SearchAuditEvents(query string) ([]*AuditEvent, error) {
	pattern := "%" + query + "%"
	rows, err := s.db.Query(
		`SELECT id, actor, target_user, action, reason, ip_address, created_at
		 FROM audit_events
		 WHERE actor ILIKE $1
		    OR target_user ILIKE $1
		    OR action ILIKE $1
		    OR COALESCE(reason, '') ILIKE $1
		    OR COALESCE(ip_address, '') ILIKE $1
		 ORDER BY created_at DESC, id DESC`,
		pattern,
	)
	if err != nil {
		return nil, fmt.Errorf("search audit events: %w", err)
	}
	defer rows.Close()

	return scanAuditEvents(rows)
}

func (s *PostgresStore) ListAuditEventsPaged(limit, offset int) ([]*AuditEvent, error) {
	rows, err := s.db.Query(
		`SELECT id, actor, target_user, action, reason, ip_address, created_at
		 FROM audit_events
		 ORDER BY created_at DESC, id DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list audit events paged: %w", err)
	}
	defer rows.Close()

	return scanAuditEvents(rows)
}

func (s *PostgresStore) SearchAuditEventsPaged(query string, limit, offset int) ([]*AuditEvent, error) {
	pattern := "%" + query + "%"
	rows, err := s.db.Query(
		`SELECT id, actor, target_user, action, reason, ip_address, created_at
		 FROM audit_events
		 WHERE actor ILIKE $1
		    OR target_user ILIKE $1
		    OR action ILIKE $1
		    OR COALESCE(reason, '') ILIKE $1
		    OR COALESCE(ip_address, '') ILIKE $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2 OFFSET $3`,
		pattern, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("search audit events paged: %w", err)
	}
	defer rows.Close()

	return scanAuditEvents(rows)
}

func (s *PostgresStore) CountAuditEvents() (int, error) {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM audit_events`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count audit events: %w", err)
	}
	return count, nil
}

func (s *PostgresStore) CountSearchAuditEvents(query string) (int, error) {
	pattern := "%" + query + "%"
	var count int
	if err := s.db.QueryRow(
		`SELECT COUNT(*)
		 FROM audit_events
		 WHERE actor ILIKE $1
		    OR target_user ILIKE $1
		    OR action ILIKE $1
		    OR COALESCE(reason, '') ILIKE $1
		    OR COALESCE(ip_address, '') ILIKE $1`,
		pattern,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count audit events search: %w", err)
	}
	return count, nil
}

func (s *PostgresStore) decryptSecret(secret *Secret) (*Secret, error) {
	if s.codec == nil {
		return nil, ErrMissingCodec
	}
	plaintext, err := s.codec.Decrypt(secret.Secret)
	if err != nil {
		return nil, err
	}
	clone := *secret
	clone.Secret = plaintext
	return &clone, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRequest(row scanner) (*Request, error) {
	var request Request
	var approved sql.NullBool
	var authUser sql.NullString
	var reasonApproval sql.NullString
	var dateApproved sql.NullTime
	if err := row.Scan(
		&request.ID,
		&request.SecretID,
		&request.RequestingUser,
		&approved,
		&authUser,
		&request.ReasonForRequest,
		&reasonApproval,
		&request.DateRequested,
		&dateApproved,
		&request.Current,
	); err != nil {
		return nil, err
	}

	if approved.Valid {
		value := approved.Bool
		request.Approved = &value
	}
	if authUser.Valid {
		request.AuthUser = authUser.String
	}
	if reasonApproval.Valid {
		request.ReasonForApproval = reasonApproval.String
	}
	if dateApproved.Valid {
		request.DateApproved = &dateApproved.Time
	}

	return &request, nil
}

func scanAuditEvents(rows *sql.Rows) ([]*AuditEvent, error) {
	events := make([]*AuditEvent, 0)
	for rows.Next() {
		var event AuditEvent
		var reason sql.NullString
		var ipAddress sql.NullString
		if err := rows.Scan(&event.ID, &event.Actor, &event.TargetUser, &event.Action, &reason, &ipAddress, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		if reason.Valid {
			event.Reason = reason.String
		}
		if ipAddress.Valid {
			event.IPAddress = ipAddress.String
		}
		events = append(events, &event)
	}
	return events, rows.Err()
}

func nullableString(value string) sql.NullString {
	if strings.TrimSpace(value) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func (s *PostgresStore) IsEmpty() (bool, error) {
	tables := []string{"computers", "secrets", "requests", "users"}
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := s.db.QueryRow(query).Scan(&count); err != nil {
			return false, fmt.Errorf("count %s: %w", table, err)
		}
		if count > 0 {
			return false, nil
		}
	}
	return true, nil
}

func (s *PostgresStore) ImportComputer(id int, serial, username, computerName string, lastCheckin time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO computers (id, serial, username, computername, last_checkin) VALUES ($1, $2, $3, $4, $5)`,
		id, serial, username, computerName, lastCheckin,
	)
	if err != nil {
		return fmt.Errorf("import computer: %w", err)
	}
	return nil
}

func (s *PostgresStore) ImportSecret(id, computerID int, secretType, encryptedSecret string, dateEscrowed time.Time, rotationRequired bool) error {
	_, err := s.db.Exec(
		`INSERT INTO secrets (id, computer_id, secret_type, secret, date_escrowed, rotation_required) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, computerID, secretType, encryptedSecret, dateEscrowed, rotationRequired,
	)
	if err != nil {
		return fmt.Errorf("import secret: %w", err)
	}
	return nil
}

func (s *PostgresStore) ImportRequest(id, secretID int, requestingUser string, approved *bool, authUser, reasonForRequest, reasonForApproval string, dateRequested time.Time, dateApproved *time.Time, current bool) error {
	var approvedValue sql.NullBool
	if approved != nil {
		approvedValue.Valid = true
		approvedValue.Bool = *approved
	}
	var dateApprovedValue sql.NullTime
	if dateApproved != nil {
		dateApprovedValue.Valid = true
		dateApprovedValue.Time = *dateApproved
	}
	_, err := s.db.Exec(
		`INSERT INTO requests (id, secret_id, requesting_user, approved, auth_user, reason_for_request, reason_for_approval, date_requested, date_approved, current) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		id, secretID, requestingUser, approvedValue, nullableString(authUser), reasonForRequest, nullableString(reasonForApproval), dateRequested, dateApprovedValue, current,
	)
	if err != nil {
		return fmt.Errorf("import request: %w", err)
	}
	return nil
}

func (s *PostgresStore) ImportUser(id int, username, passwordHash string, isStaff, canApprove, localLoginEnabled, mustResetPassword bool, authSource string) error {
	_, err := s.db.Exec(
		`INSERT INTO users (id, username, password_hash, is_staff, can_approve, local_login_enabled, must_reset_password, auth_source) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, username, nullableString(passwordHash), isStaff, canApprove, localLoginEnabled, mustResetPassword, authSource,
	)
	if err != nil {
		return fmt.Errorf("import user: %w", err)
	}
	return nil
}
