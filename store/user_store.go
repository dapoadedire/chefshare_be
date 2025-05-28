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
	ID             int      `json:"id"`
	Username       string   `json:"username"`
	Email          string   `json:"email"`
	PasswordHash   password `json:"password_hash"`
	Bio            string   `json:"bio"`
	FirstName      string   `json:"first_name"`
	LastName       string   `json:"last_name"`
	ProfilePicture string   `json:"profile_picture"`
	LastLogin      *string  `json:"last_login"`
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
	query := `INSERT INTO users(username, email, password_hash, bio, first_name, last_name, profile_picture)
	VALUES($1,$2,$3,$4,$5,$6,$7)
	RETURNING id, created_at, updated_at
	`
	err := s.db.QueryRow(query, user.Username, user.Email, user.PasswordHash.hash, user.Bio, user.FirstName, user.LastName, user.ProfilePicture).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresUserStore) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, bio, first_name, last_name, profile_picture, 
		       last_login, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}
	var passwordHash []byte

	err := s.db.QueryRow(query, email).Scan(
		&user.ID,
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
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}

	user.PasswordHash.hash = passwordHash
	return user, nil
}

func (s *PostgresUserStore) GetUserByID(id int64) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, bio, first_name, last_name, profile_picture, 
		       last_login, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	var passwordHash []byte

	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
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
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int64) (*User, error)
	UpdatePassword(userID int64, newPassword string) error
}

// UpdatePassword updates a user's password
func (s *PostgresUserStore) UpdatePassword(userID int64, newPassword string) error {
	// Create a temporary password struct to generate the hash
	var pass password
	if err := pass.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the password in the database
	query := `
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	_, err := s.db.Exec(query, pass.hash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{
		db: db,
	}
}
