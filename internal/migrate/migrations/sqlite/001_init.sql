CREATE TABLE IF NOT EXISTS computers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    serial TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    computername TEXT NOT NULL,
    last_checkin TIMESTAMP
);

CREATE TABLE IF NOT EXISTS secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    computer_id INTEGER NOT NULL REFERENCES computers(id) ON DELETE CASCADE,
    secret TEXT NOT NULL,
    secret_type TEXT NOT NULL,
    date_escrowed TIMESTAMP NOT NULL,
    rotation_required BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    secret_id INTEGER NOT NULL REFERENCES secrets(id) ON DELETE RESTRICT,
    requesting_user TEXT NOT NULL,
    approved BOOLEAN NULL,
    auth_user TEXT NULL,
    reason_for_request TEXT NOT NULL,
    reason_for_approval TEXT NULL,
    date_requested TIMESTAMP NOT NULL,
    date_approved TIMESTAMP NULL,
    current BOOLEAN NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NULL,
    is_staff BOOLEAN NOT NULL DEFAULT 0,
    can_approve BOOLEAN NOT NULL DEFAULT 0,
    local_login_enabled BOOLEAN NOT NULL DEFAULT 0,
    must_reset_password BOOLEAN NOT NULL DEFAULT 0,
    auth_source TEXT NOT NULL DEFAULT 'local'
);
