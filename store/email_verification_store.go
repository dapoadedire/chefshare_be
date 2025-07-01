package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

type EmailVerificationToken struct {
	ID        int64
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type EmailVerificationStore interface {
	CreateVerificationToken(userID string, expiryDuration time.Duration) (*EmailVerificationToken, error)
	GetVerificationTokenByToken(token string) (*EmailVerificationToken, error)
	DeleteToken(tokenID int64) error
	DeleteUserTokens(userID string) (int64, error)
	DeleteExpiredTokens() (int64, error)
}

type PostgresEmailVerificationStore struct {
	db *sql.DB
}

func NewPostgresEmailVerificationStore(db *sql.DB) *PostgresEmailVerificationStore {
	return &PostgresEmailVerificationStore{
		db: db,
	}
}

// GenerateVerificationToken creates a secure random token for email verification
func generateVerificationToken() (string, error) {
	// Generate 32 bytes of random data
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as URL-safe base64 and remove any padding
	token := base64.URLEncoding.EncodeToString(b)
	token = token[:42] // Truncate to reasonable length

	return token, nil
}

// CreateVerificationToken creates a new email verification token
func (s *PostgresEmailVerificationStore) CreateVerificationToken(userID string, expiryDuration time.Duration) (*EmailVerificationToken, error) {
	// First, invalidate any existing tokens for this user
	_, err := s.DeleteUserTokens(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	// Generate a new token
	token, err := generateVerificationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(expiryDuration)

	verificationToken := &EmailVerificationToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	query := `
		INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err = s.db.QueryRow(query, verificationToken.UserID, verificationToken.Token,
		verificationToken.ExpiresAt).Scan(
		&verificationToken.ID, &verificationToken.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create verification token: %w", err)
	}

	return verificationToken, nil
}

// GetVerificationTokenByToken retrieves a token by its value
func (s *PostgresEmailVerificationStore) GetVerificationTokenByToken(token string) (*EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM email_verification_tokens
		WHERE token = $1
	`

	verificationToken := &EmailVerificationToken{}
	err := s.db.QueryRow(query, token).Scan(
		&verificationToken.ID,
		&verificationToken.UserID,
		&verificationToken.Token,
		&verificationToken.ExpiresAt,
		&verificationToken.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Token not found
		}
		return nil, fmt.Errorf("failed to get verification token: %w", err)
	}

	return verificationToken, nil
}

// DeleteToken removes a token by ID
func (s *PostgresEmailVerificationStore) DeleteToken(tokenID int64) error {
	query := `DELETE FROM email_verification_tokens WHERE id = $1`

	_, err := s.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

// DeleteExpiredTokens removes all expired tokens
func (s *PostgresEmailVerificationStore) DeleteExpiredTokens() (int64, error) {
	query := `DELETE FROM email_verification_tokens WHERE expires_at < $1`

	result, err := s.db.Exec(query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// DeleteUserTokens removes all tokens for a specific user
func (s *PostgresEmailVerificationStore) DeleteUserTokens(userID string) (int64, error) {
	query := `DELETE FROM email_verification_tokens WHERE user_id = $1`

	result, err := s.db.Exec(query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
