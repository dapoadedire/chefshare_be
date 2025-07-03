-- +goose Up
-- +goose StatementBegin
-- Create a table for blacklisted tokens
CREATE TABLE IF NOT EXISTS blacklisted_tokens (
    id SERIAL PRIMARY KEY,
    token VARCHAR(1000) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create an index on the token for fast lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_blacklisted_tokens_token ON blacklisted_tokens(token);

-- Create an index on the expiry date for efficient cleanup
CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS blacklisted_tokens;
-- +goose StatementEnd
