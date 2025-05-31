package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionStore interface {
	CreateSession(userID int64, duration time.Duration) (*Session, error)
	GetSessionByToken(token string) (*Session, error)
	DeleteSession(token string) error
	DeleteExpiredSessions() (int64, error)
	DeleteUserSessions(userID int64) (int64, error)
}

type PostgresSessionStore struct {
	db *sql.DB
}

func NewPostgresSessionStore(db *sql.DB) *PostgresSessionStore {
	return &PostgresSessionStore{
		db: db,
	}
}

// CreateSession creates a new session for the given user
func (s *PostgresSessionStore) CreateSession(userID int64, duration time.Duration) (*Session, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(duration)

	session := &Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	query := `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`

	err := s.db.QueryRow(query, session.Token, session.UserID, session.ExpiresAt).Scan(&session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSessionByToken retrieves a session by its token
func (s *PostgresSessionStore) GetSessionByToken(token string) (*Session, error) {
	query := `
		SELECT token, user_id, expires_at, created_at
		FROM sessions
		WHERE token = $1
	`

	session := &Session{}
	err := s.db.QueryRow(query, token).Scan(
		&session.Token,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// DeleteSession removes a session by its token
func (s *PostgresSessionStore) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`

	_, err := s.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteExpiredSessions removes all expired sessions
func (s *PostgresSessionStore) DeleteExpiredSessions() (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < $1`

	result, err := s.db.Exec(query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// DeleteUserSessions removes all sessions for a specific user
func (s *PostgresSessionStore) DeleteUserSessions(userID int64) (int64, error) {
	query := `DELETE FROM sessions WHERE user_id = $1`

	result, err := s.db.Exec(query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
