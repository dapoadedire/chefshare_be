package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID        int64     `json:"id"`
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	IssuedAt  time.Time `json:"issued_at"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// RefreshTokenStore defines the interface for refresh token operations
type RefreshTokenStore interface {
	CreateRefreshToken(userID string, duration time.Duration, ipAddress, userAgent string) (*RefreshToken, error)
	CreateRefreshTokenWithTransaction(userID string, duration time.Duration, ipAddress, userAgent string, tx *sql.Tx) (*RefreshToken, error)
	GetRefreshToken(token string) (*RefreshToken, error)
	RevokeRefreshToken(token string) error
	RevokeAllUserRefreshTokens(userID string) (int64, error)
	DeleteExpiredRefreshTokens() (int64, error)
}

// PostgresRefreshTokenStore implements the RefreshTokenStore interface using PostgreSQL
type PostgresRefreshTokenStore struct {
	db *sql.DB
}

// NewPostgresRefreshTokenStore creates a new PostgresRefreshTokenStore
func NewPostgresRefreshTokenStore(db *sql.DB) *PostgresRefreshTokenStore {
	return &PostgresRefreshTokenStore{
		db: db,
	}
}

// CreateRefreshToken creates a new refresh token for the given user
func (s *PostgresRefreshTokenStore) CreateRefreshToken(userID string, duration time.Duration, ipAddress, userAgent string) (*RefreshToken, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(duration)

	refreshToken := &RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Revoked:   false,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, issued_at
	`

	err := s.db.QueryRow(
		query,
		refreshToken.Token,
		refreshToken.UserID,
		refreshToken.ExpiresAt,
		refreshToken.IPAddress,
		refreshToken.UserAgent,
	).Scan(&refreshToken.ID, &refreshToken.IssuedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return refreshToken, nil
}

// CreateRefreshTokenWithTransaction creates a new refresh token for the given user within a transaction
func (s *PostgresRefreshTokenStore) CreateRefreshTokenWithTransaction(userID string, duration time.Duration, ipAddress, userAgent string, tx *sql.Tx) (*RefreshToken, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(duration)

	refreshToken := &RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Revoked:   false,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, issued_at
	`

	err := tx.QueryRow(
		query,
		refreshToken.Token,
		refreshToken.UserID,
		refreshToken.ExpiresAt,
		refreshToken.IPAddress,
		refreshToken.UserAgent,
	).Scan(&refreshToken.ID, &refreshToken.IssuedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token in transaction: %w", err)
	}

	return refreshToken, nil
}

// GetRefreshToken retrieves a refresh token by its token string
func (s *PostgresRefreshTokenStore) GetRefreshToken(token string) (*RefreshToken, error) {
	query := `
		SELECT id, token, user_id, expires_at, revoked, issued_at, ip_address, user_agent
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > $2
	`

	refreshToken := &RefreshToken{}
	err := s.db.QueryRow(query, token, time.Now()).Scan(
		&refreshToken.ID,
		&refreshToken.Token,
		&refreshToken.UserID,
		&refreshToken.ExpiresAt,
		&refreshToken.Revoked,
		&refreshToken.IssuedAt,
		&refreshToken.IPAddress,
		&refreshToken.UserAgent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Token not found or expired
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Additional check for already revoked tokens (should be caught by DELETE approach now)
	if refreshToken.Revoked {
		return nil, fmt.Errorf("token has been revoked")
	}

	return refreshToken, nil
}

// RevokeRefreshToken deletes a refresh token from the database
func (s *PostgresRefreshTokenStore) RevokeRefreshToken(token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`

	result, err := s.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// RevokeAllUserRefreshTokens deletes all refresh tokens for a specific user
func (s *PostgresRefreshTokenStore) RevokeAllUserRefreshTokens(userID string) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`

	result, err := s.db.Exec(query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// DeleteExpiredRefreshTokens removes all expired refresh tokens
func (s *PostgresRefreshTokenStore) DeleteExpiredRefreshTokens() (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`

	result, err := s.db.Exec(query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
