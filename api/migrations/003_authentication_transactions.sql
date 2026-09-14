CREATE TABLE IF NOT EXISTS authentication_login_transactions (
    state_hash CHAR(64) PRIMARY KEY,
    nonce TEXT NOT NULL,
    verifier TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_authentication_login_transactions_expiry
    ON authentication_login_transactions (expires_at);

ALTER TABLE authentication_sessions
    ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 1;
