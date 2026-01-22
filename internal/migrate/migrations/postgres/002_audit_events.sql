CREATE TABLE IF NOT EXISTS audit_events (
    id SERIAL PRIMARY KEY,
    actor TEXT NOT NULL,
    target_user TEXT NOT NULL,
    action TEXT NOT NULL,
    reason TEXT NULL,
    ip_address TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
