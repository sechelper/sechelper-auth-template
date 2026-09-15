CREATE TABLE IF NOT EXISTS platform_users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS external_identities (
    issuer TEXT NOT NULL,
    identity_subject TEXT NOT NULL,
    platform_user_uuid UUID NOT NULL REFERENCES platform_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (issuer, identity_subject)
);

CREATE INDEX IF NOT EXISTS idx_external_identities_platform_user
    ON external_identities (platform_user_uuid);

ALTER TABLE authentication_sessions
    ADD COLUMN IF NOT EXISTS platform_user_uuid UUID REFERENCES platform_users(id);

CREATE INDEX IF NOT EXISTS idx_authentication_sessions_platform_user_created
    ON authentication_sessions (platform_user_uuid, created_at DESC, id DESC)
    WHERE platform_user_uuid IS NOT NULL;
