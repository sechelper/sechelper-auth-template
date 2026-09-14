-- Forward-only schema for the durable Session Store.
-- The first implementation uses memory for local development; this schema is the production boundary.
CREATE TABLE IF NOT EXISTS authentication_sessions (
    id TEXT PRIMARY KEY,
    subject TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    application_code TEXT NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    refresh_token_ciphertext TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_authentication_sessions_subject_created
    ON authentication_sessions (subject, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_authentication_sessions_expiry
    ON authentication_sessions (expires_at);

CREATE TABLE IF NOT EXISTS authorization_manifest_state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    application_code TEXT NOT NULL,
    manifest_version BIGINT NOT NULL,
    content_hash TEXT NOT NULL,
    canonical_json TEXT NOT NULL DEFAULT '{}',
    sync_id TEXT,
    status TEXT NOT NULL,
    server_revision BIGINT NOT NULL DEFAULT 0,
    accepted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL
);
