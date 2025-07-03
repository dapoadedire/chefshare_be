package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dapoadedire/chefshare_be/store"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds configuration for the JWT service
type JWTConfig struct {
	AccessTokenSecret      string
	RefreshTokenSecret     string
	AccessTokenDuration    time.Duration
	RefreshTokenDuration   time.Duration
	AccessTokenCookieName  string
	RefreshTokenCookieName string
}

// DefaultJWTConfig returns a default JWT configuration
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		AccessTokenSecret:      getEnvOrDefault("JWT_ACCESS_SECRET", "default_access_secret_change_me_in_production"),
		RefreshTokenSecret:     getEnvOrDefault("JWT_REFRESH_SECRET", "default_refresh_secret_change_me_in_production"),
		AccessTokenDuration:    15 * time.Minute,
		RefreshTokenDuration:   7 * 24 * time.Hour, // 7 days
		AccessTokenCookieName:  "access_token",
		RefreshTokenCookieName: "refresh_token",
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// CustomClaims extends standard claims with custom user information
type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation
type JWTService struct {
	config              JWTConfig
	refreshTokenStore   store.RefreshTokenStore
	tokenBlacklistStore store.TokenBlacklistStore
	userStore           store.UserStore
}

// GetConfig returns the JWTService configuration
func (s *JWTService) GetConfig() JWTConfig {
	return s.config
}

// NewJWTService creates a new JWT service with the given configuration
func NewJWTService(config JWTConfig, refreshTokenStore store.RefreshTokenStore, userStore store.UserStore, tokenBlacklistStore store.TokenBlacklistStore) *JWTService {
	return &JWTService{
		config:              config,
		refreshTokenStore:   refreshTokenStore,
		userStore:           userStore,
		tokenBlacklistStore: tokenBlacklistStore,
	}
}

// GenerateTokenPair creates both access and refresh tokens for a user
func (s *JWTService) GenerateTokenPair(user *store.User, ipAddress, userAgent string) (string, *store.RefreshToken, error) {
	// Generate access token with short expiry
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Store refresh token in database
	refreshToken, err := s.refreshTokenStore.CreateRefreshToken(
		user.UserID,
		s.config.RefreshTokenDuration,
		ipAddress,
		userAgent,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateTokenPairWithTransaction creates both access and refresh tokens for a user within a transaction
func (s *JWTService) GenerateTokenPairWithTransaction(user *store.User, ipAddress, userAgent string, tx *sql.Tx) (string, *store.RefreshToken, error) {
	// Generate access token with short expiry
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Store refresh token in database using the transaction
	refreshToken, err := s.refreshTokenStore.CreateRefreshTokenWithTransaction(
		user.UserID,
		s.config.RefreshTokenDuration,
		ipAddress,
		userAgent,
		tx,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create refresh token in transaction: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken creates a new JWT access token
func (s *JWTService) GenerateAccessToken(user *store.User) (string, error) {
	// Set token expiry time
	expirationTime := time.Now().Add(s.config.AccessTokenDuration)

	// Create claims with user information
	claims := &CustomClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "chefshare_api",
			Subject:   user.UserID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(s.config.AccessTokenSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// RefreshAccessToken validates a refresh token, generates a new access token, and rotates the refresh token
func (s *JWTService) RefreshAccessToken(refreshTokenString string) (string, *store.RefreshToken, error) {
	// Get refresh token from database
	refreshToken, err := s.refreshTokenStore.GetRefreshToken(refreshTokenString)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Check if token exists
	if refreshToken == nil {
		return "", nil, fmt.Errorf("invalid refresh token")
	}

	// Note: Token expiration check is now done in the GetRefreshToken method via SQL query

	// Get user from the database using the UserStore
	user, err := s.userStore.GetUserByID(refreshToken.UserID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return "", nil, fmt.Errorf("user not found")
	}

	// Start a database transaction
	db := s.userStore.DB()
	tx, err := db.Begin()
	if err != nil {
		return "", nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer a function to handle transaction completion based on success or failure
	defer func() {
		if err != nil {
			// If we encountered an error, roll back the transaction
			if txErr := tx.Rollback(); txErr != nil {
				log.Printf("Failed to rollback transaction: %v", txErr)
			}
		}
	}()

	// Revoke the current refresh token
	err = s.refreshTokenStore.RevokeRefreshToken(refreshTokenString)
	if err != nil {
		return "", nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Generate a new refresh token
	ipAddress := refreshToken.IPAddress
	userAgent := refreshToken.UserAgent

	// Generate new access token and refresh token
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Create new refresh token
	newRefreshToken, err := s.refreshTokenStore.CreateRefreshToken(
		user.UserID,
		s.config.RefreshTokenDuration,
		ipAddress,
		userAgent,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return "", nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// ValidateAccessToken validates the provided JWT access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	// First, check if token is blacklisted
	isBlacklisted, err := s.tokenBlacklistStore.IsBlacklisted(tokenString)
	if err != nil {
		// Log error but continue validation
		log.Printf("Error checking token blacklist: %v", err)
	}

	if isBlacklisted {
		return nil, fmt.Errorf("token is revoked")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.AccessTokenSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (s *JWTService) RevokeRefreshToken(tokenString string) error {
	return s.refreshTokenStore.RevokeRefreshToken(tokenString)
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a specific user
func (s *JWTService) RevokeAllUserRefreshTokens(userID string) (int64, error) {
	return s.refreshTokenStore.RevokeAllUserRefreshTokens(userID)
}

// BlacklistAccessToken adds an access token to the blacklist
func (s *JWTService) BlacklistAccessToken(tokenString string) error {
	// Parse the token to get the expiry time
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.AccessTokenSecret), nil
	})

	var expiresAt time.Time

	// Even if token is invalid, we'll blacklist it anyway with a reasonable expiry
	if err == nil && token.Valid {
		if claims, ok := token.Claims.(*CustomClaims); ok {
			expiresAt = claims.ExpiresAt.Time
		}
	} else {
		// If we can't parse the token, blacklist it with a default expiry of 24 hours
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	// Add token to blacklist
	return s.tokenBlacklistStore.BlacklistToken(tokenString, expiresAt)
}
