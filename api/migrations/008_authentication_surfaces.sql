ALTER TABLE authentication_sessions
    ADD COLUMN IF NOT EXISTS surface TEXT NOT NULL DEFAULT 'public';

ALTER TABLE authentication_login_transactions
    ADD COLUMN IF NOT EXISTS surface TEXT NOT NULL DEFAULT 'public';

ALTER TABLE authentication_login_transactions
    ADD COLUMN IF NOT EXISTS return_to TEXT NOT NULL DEFAULT '/';

ALTER TABLE authentication_sessions
    DROP CONSTRAINT IF EXISTS authentication_sessions_surface_check;

ALTER TABLE authentication_sessions
    ADD CONSTRAINT authentication_sessions_surface_check CHECK (surface IN ('public', 'admin'));

ALTER TABLE authentication_login_transactions
    DROP CONSTRAINT IF EXISTS authentication_login_transactions_surface_check;

ALTER TABLE authentication_login_transactions
    ADD CONSTRAINT authentication_login_transactions_surface_check CHECK (surface IN ('public', 'admin'));

CREATE INDEX IF NOT EXISTS idx_authentication_sessions_surface_expiry
    ON authentication_sessions (surface, expires_at);
