CREATE TABLE IF NOT EXISTS computers (
    id SERIAL PRIMARY KEY,
    serial TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    computername TEXT NOT NULL,
    last_checkin TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS secrets (
    id SERIAL PRIMARY KEY,
    computer_id INTEGER NOT NULL REFERENCES computers(id) ON DELETE CASCADE,
    secret TEXT NOT NULL,
    secret_type TEXT NOT NULL,
    date_escrowed TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotation_required BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS requests (
    id SERIAL PRIMARY KEY,
    secret_id INTEGER NOT NULL REFERENCES secrets(id) ON DELETE RESTRICT,
    requesting_user TEXT NOT NULL,
    approved BOOLEAN NULL,
    auth_user TEXT NULL,
    reason_for_request TEXT NOT NULL,
    reason_for_approval TEXT NULL,
    date_requested TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    date_approved TIMESTAMPTZ NULL,
    current BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NULL,
    is_staff BOOLEAN NOT NULL DEFAULT FALSE,
    can_approve BOOLEAN NOT NULL DEFAULT FALSE,
    local_login_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    must_reset_password BOOLEAN NOT NULL DEFAULT FALSE,
    auth_source TEXT NOT NULL DEFAULT 'local'
);
