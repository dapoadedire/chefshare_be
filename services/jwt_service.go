package services

import (
	"fmt"
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
	config            JWTConfig
	refreshTokenStore store.RefreshTokenStore
	userStore         store.UserStore
}

// GetConfig returns the JWTService configuration
func (s *JWTService) GetConfig() JWTConfig {
	return s.config
}

// NewJWTService creates a new JWT service with the given configuration
func NewJWTService(config JWTConfig, refreshTokenStore store.RefreshTokenStore, userStore store.UserStore) *JWTService {
	return &JWTService{
		config:            config,
		refreshTokenStore: refreshTokenStore,
		userStore:         userStore,
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

// RefreshAccessToken validates a refresh token and generates a new access token
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

	// Check if token is revoked
	if refreshToken.Revoked {
		return "", nil, fmt.Errorf("refresh token has been revoked")
	}

	// Check if token is expired
	if refreshToken.ExpiresAt.Before(time.Now()) {
		return "", nil, fmt.Errorf("refresh token has expired")
	}

	// Get user from the database using the UserStore
	user, err := s.userStore.GetUserByID(refreshToken.UserID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return "", nil, fmt.Errorf("user not found")
	}

	// Generate new access token
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateAccessToken validates the provided JWT access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
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