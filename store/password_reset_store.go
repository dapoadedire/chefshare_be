package store

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

type PasswordResetToken struct {
	ID        int64
	UserID    string
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

type PasswordResetStore interface {
	CreatePasswordResetToken(userID string, expiryDuration time.Duration) (*PasswordResetToken, error)
	GetPasswordResetTokenByToken(token string) (*PasswordResetToken, error)
	MarkTokenAsUsed(tokenID int64) error
	DeleteExpiredTokens() (int64, error)
	DeleteUserTokens(userID string) (int64, error)
}

type PostgresPasswordResetStore struct {
	db *sql.DB
}

func NewPostgresPasswordResetStore(db *sql.DB) *PostgresPasswordResetStore {
	return &PostgresPasswordResetStore{
		db: db,
	}
}

// GenerateOTP generates a 6-digit OTP
func generateOTP() string {
	// Set up a random number generator with the current time as seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random 6-digit number
	otp := r.Intn(900000) + 100000 // This ensures a 6-digit number (100000 to 999999)

	return fmt.Sprintf("%06d", otp)
}

// CreatePasswordResetToken creates a new password reset token for the given user
func (s *PostgresPasswordResetStore) CreatePasswordResetToken(userID string, expiryDuration time.Duration) (*PasswordResetToken, error) {
	// First, invalidate any existing tokens for this user
	_, err := s.DeleteUserTokens(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	// Generate a new OTP
	token := generateOTP()
	expiresAt := time.Now().Add(expiryDuration)

	passwordResetToken := &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, used)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err = s.db.QueryRow(query, passwordResetToken.UserID, passwordResetToken.Token,
		passwordResetToken.ExpiresAt, passwordResetToken.Used).Scan(
		&passwordResetToken.ID, &passwordResetToken.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create password reset token: %w", err)
	}

	return passwordResetToken, nil
}

// GetPasswordResetTokenByToken retrieves a token by its value
func (s *PostgresPasswordResetStore) GetPasswordResetTokenByToken(token string) (*PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	passwordResetToken := &PasswordResetToken{}
	err := s.db.QueryRow(query, token).Scan(
		&passwordResetToken.ID,
		&passwordResetToken.UserID,
		&passwordResetToken.Token,
		&passwordResetToken.ExpiresAt,
		&passwordResetToken.Used,
		&passwordResetToken.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Token not found
		}
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}

	return passwordResetToken, nil
}

// MarkTokenAsUsed marks a token as used
func (s *PostgresPasswordResetStore) MarkTokenAsUsed(tokenID int64) error {
	query := `UPDATE password_reset_tokens SET used = true WHERE id = $1`

	_, err := s.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// DeleteExpiredTokens removes all expired tokens
func (s *PostgresPasswordResetStore) DeleteExpiredTokens() (int64, error) {
	query := `DELETE FROM password_reset_tokens WHERE expires_at < $1`

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
func (s *PostgresPasswordResetStore) DeleteUserTokens(userID string) (int64, error) {
	query := `DELETE FROM password_reset_tokens WHERE user_id = $1`

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
