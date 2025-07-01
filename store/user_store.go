package store

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type password struct {
	hash      []byte
	plainText *string
}

type User struct {
	UserID         string   `json:"user_id"`
	Username       string   `json:"username"`
	Email          string   `json:"email"`
	PasswordHash   password `json:"password_hash"`
	Bio            string   `json:"bio"`
	FirstName      string   `json:"first_name"`
	LastName       string   `json:"last_name"`
	ProfilePicture string   `json:"profile_picture"`
	LastLogin      *string  `json:"last_login"`
	EmailVerified  bool     `json:"email_verified"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

func (password *password) SetPassword(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	password.plainText = &plaintextPassword
	password.hash = hash
	return nil
}

func (password *password) CheckPassword(plaintextPassword string) error {
	return bcrypt.CompareHashAndPassword(password.hash, []byte(plaintextPassword))
}

type PostgresUserStore struct {
	db *sql.DB
}

func (s *PostgresUserStore) CreateUser(user *User) error {
	query := `INSERT INTO users(user_id, username, email, password_hash, bio, first_name, last_name, profile_picture)
	VALUES($1,$2,$3,$4,$5,$6,$7, $8)
	RETURNING user_id, created_at, updated_at
	`
	err := s.db.QueryRow(query, user.UserID, user.Username, user.Email, user.PasswordHash.hash, user.Bio, user.FirstName, user.LastName, user.ProfilePicture).Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresUserStore) CreateUserWithTransaction(user *User, tx *sql.Tx) error {
	query := `INSERT INTO users(user_id, username, email, password_hash, bio, first_name, last_name, profile_picture)
	VALUES($1,$2,$3,$4,$5,$6,$7, $8)
	RETURNING user_id, created_at, updated_at
	`
	err := tx.QueryRow(query, user.UserID, user.Username, user.Email, user.PasswordHash.hash, user.Bio, user.FirstName, user.LastName, user.ProfilePicture).Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresUserStore) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT user_id, username, email, password_hash, bio, first_name, last_name, profile_picture, 
		       last_login, email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}
	var passwordHash []byte

	err := s.db.QueryRow(query, email).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.Bio,
		&user.FirstName,
		&user.LastName,
		&user.ProfilePicture,
		&user.LastLogin,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}

	user.PasswordHash.hash = passwordHash
	return user, nil
}

func (s *PostgresUserStore) GetUserByID(userID string) (*User, error) {
	query := `
		SELECT user_id, username, email, password_hash, bio, first_name, last_name, profile_picture, 
		       last_login, email_verified, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	user := &User{}
	var passwordHash []byte

	err := s.db.QueryRow(query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.Bio,
		&user.FirstName,
		&user.LastName,
		&user.ProfilePicture,
		&user.LastLogin,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}

	user.PasswordHash.hash = passwordHash
	return user, nil
}

type UserStore interface {
	CreateUser(user *User) error
	CreateUserWithTransaction(user *User, tx *sql.Tx) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(userID string) (*User, error)
	UpdatePassword(userID string, newPassword string) error
	UpdateUser(userID string, updates map[string]interface{}) (*User, error)
	UpdateLastLogin(userID string) error
	IsUsernameTaken(username string, excludeUserID string) (bool, error)
	SetEmailVerified(userID string, verified bool) error
	DB() *sql.DB
}

// UpdatePassword updates a user's password
func (s *PostgresUserStore) UpdatePassword(userID string, newPassword string) error {
	// Create a temporary password struct to generate the hash
	var pass password
	if err := pass.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the password in the database
	query := `
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`

	_, err := s.db.Exec(query, pass.hash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdateUser updates user profile information and returns the updated user
func (s *PostgresUserStore) UpdateUser(userID string, updates map[string]interface{}) (*User, error) {
	if len(updates) == 0 {
		// If there are no updates, just return the current user data
		return s.GetUserByID(userID)
	}

	// Build the dynamic query
	query := "UPDATE users SET "
	params := make([]interface{}, 0, len(updates))
	i := 1

	for field, value := range updates {
		if i > 1 {
			query += ", "
		}
		query += field + " = $" + fmt.Sprint(i)
		params = append(params, value)
		i++
	}

	// Add RETURNING clause to get the updated user data
	query += " WHERE user_id = $" + fmt.Sprint(i) + " RETURNING user_id, username, email, password_hash, bio, first_name, last_name, profile_picture, last_login, created_at, updated_at"
	params = append(params, userID)

	// Execute the query and scan results directly into a User object
	user := &User{}
	var passwordHash []byte

	err := s.db.QueryRow(query, params...).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.Bio,
		&user.FirstName,
		&user.LastName,
		&user.ProfilePicture,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user.PasswordHash.hash = passwordHash
	return user, nil
}

// UpdateLastLogin updates the last_login timestamp for a user to the current time
func (s *PostgresUserStore) UpdateLastLogin(userID string) error {
	query := `
		UPDATE users 
		SET last_login = CURRENT_TIMESTAMP 
		WHERE user_id = $1
	`

	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last_login: %w", err)
	}

	return nil
}

// DB returns the underlying database connection
// This is needed for more complex queries that aren't part of the standard interface
func (s *PostgresUserStore) DB() *sql.DB {
	return s.db
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{
		db: db,
	}
}

// IsUsernameTaken checks if a username is already taken by another user
func (s *PostgresUserStore) IsUsernameTaken(username string, excludeUserID string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM users 
		WHERE username = $1 AND user_id != $2
	`

	var count int
	err := s.db.QueryRow(query, username, excludeUserID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}

// SetEmailVerified updates the email_verified status for a user
func (s *PostgresUserStore) SetEmailVerified(userID string, verified bool) error {
	query := `
		UPDATE users 
		SET email_verified = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`

	_, err := s.db.Exec(query, verified, userID)
	if err != nil {
		return fmt.Errorf("failed to update email verification status: %w", err)
	}

	return nil
}
