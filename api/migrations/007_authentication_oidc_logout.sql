ALTER TABLE authentication_sessions
    ADD COLUMN IF NOT EXISTS id_token_ciphertext TEXT;
