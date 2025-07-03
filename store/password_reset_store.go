package store

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	// New method for transactional password reset
	ResetPasswordTransaction(tokenID int64, userID string, newPassword string) error
}

type PostgresPasswordResetStore struct {
	db *sql.DB
}

func NewPostgresPasswordResetStore(db *sql.DB) *PostgresPasswordResetStore {
	return &PostgresPasswordResetStore{
		db: db,
	}
}

// generateOTP generates a secure 6-digit OTP
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000)) // range: 0–899999
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

// CreatePasswordResetToken creates a new password reset token for the given user
func (s *PostgresPasswordResetStore) CreatePasswordResetToken(userID string, expiryDuration time.Duration) (*PasswordResetToken, error) {
	// First, invalidate any existing tokens for this user
	_, err := s.DeleteUserTokens(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	// Generate a new OTP
	token, err := generateOTP()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}
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

// ResetPasswordTransaction performs password reset in a single transaction
// This ensures atomicity between password update and marking the token as used
func (s *PostgresPasswordResetStore) ResetPasswordTransaction(tokenID int64, userID string, newPassword string) error {
	// Hash the password first
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Use defer with a closure to handle rollback if needed
	committed := false
	defer func() {
		if !committed {
			if rbErr := tx.Rollback(); rbErr != nil {
				// Just log the rollback error
				fmt.Printf("Error rolling back transaction: %v\n", rbErr)
			}
		}
	}()

	// 1. Update the user's password within the transaction
	_, err = tx.Exec(`
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password in transaction: %w", err)
	}

	// 2. Mark the token as used within the same transaction
	_, err = tx.Exec(`UPDATE password_reset_tokens SET used = true WHERE id = $1`, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used in transaction: %w", err)
	}

	// 3. Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}
