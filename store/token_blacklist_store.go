package store

import (
	"database/sql"
	"fmt"
	"time"
)

// BlacklistedToken represents a revoked token in the database
type BlacklistedToken struct {
	ID        int64     `json:"id"`
	Token     string    `json:"token"` // This will store the token's jti or a hash of the token
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenBlacklistStore defines the interface for blacklisted token operations
type TokenBlacklistStore interface {
	BlacklistToken(tokenString string, expiresAt time.Time) error
	IsBlacklisted(tokenString string) (bool, error)
	CleanupExpiredTokens() (int64, error)
}

// PostgresTokenBlacklistStore implements the TokenBlacklistStore interface
type PostgresTokenBlacklistStore struct {
	db *sql.DB
}

// NewPostgresTokenBlacklistStore creates a new PostgresTokenBlacklistStore
func NewPostgresTokenBlacklistStore(db *sql.DB) *PostgresTokenBlacklistStore {
	return &PostgresTokenBlacklistStore{
		db: db,
	}
}

// BlacklistToken adds a token to the blacklist
func (s *PostgresTokenBlacklistStore) BlacklistToken(tokenString string, expiresAt time.Time) error {
	query := `
		INSERT INTO blacklisted_tokens (token, expires_at)
		VALUES ($1, $2)
		ON CONFLICT (token) DO NOTHING
	`

	_, err := s.db.Exec(query, tokenString, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (s *PostgresTokenBlacklistStore) IsBlacklisted(tokenString string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM blacklisted_tokens
			WHERE token = $1 AND expires_at > $2
		)
	`

	var exists bool
	err := s.db.QueryRow(query, tokenString, time.Now()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if token is blacklisted: %w", err)
	}

	return exists, nil
}

// CleanupExpiredTokens removes all expired blacklisted tokens
func (s *PostgresTokenBlacklistStore) CleanupExpiredTokens() (int64, error) {
	query := `DELETE FROM blacklisted_tokens WHERE expires_at < $1`

	result, err := s.db.Exec(query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired blacklisted tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
