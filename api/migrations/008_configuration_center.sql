CREATE TABLE IF NOT EXISTS configuration_entries (
    key TEXT PRIMARY KEY,
    description TEXT NOT NULL DEFAULT '',
    is_secret BOOLEAN NOT NULL DEFAULT TRUE,
    ciphertext TEXT NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS configuration_entries_updated_at_idx ON configuration_entries (updated_at DESC);
