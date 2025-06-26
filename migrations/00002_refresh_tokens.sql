-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    user_id VARCHAR(50) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_token_idx ON refresh_tokens (token);
CREATE INDEX IF NOT EXISTS refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);
CREATE INDEX IF NOT EXISTS refresh_tokens_revoked_idx ON refresh_tokens (revoked);

-- Comment on table and columns
COMMENT ON TABLE refresh_tokens IS 'Stores JWT refresh tokens for authentication';
COMMENT ON COLUMN refresh_tokens.token IS 'The unique refresh token string';
COMMENT ON COLUMN refresh_tokens.user_id IS 'Foreign key to users table';
COMMENT ON COLUMN refresh_tokens.expires_at IS 'When this refresh token expires';
COMMENT ON COLUMN refresh_tokens.revoked IS 'Whether this token has been explicitly revoked';
COMMENT ON COLUMN refresh_tokens.issued_at IS 'When this token was issued';
COMMENT ON COLUMN refresh_tokens.ip_address IS 'IP address of the client when token was issued';
COMMENT ON COLUMN refresh_tokens.user_agent IS 'User agent of the client when token was issued';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_tokens;
-- +goose StatementEnd