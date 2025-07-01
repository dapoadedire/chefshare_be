package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dapoadedire/chefshare_be/services"
	"github.com/dapoadedire/chefshare_be/store"
	"github.com/dapoadedire/chefshare_be/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// DefaultSessionDuration is the default duration for sessions (7 days)
	DefaultSessionDuration = 7 * 24 * time.Hour

	// EmailVerificationTokenExpiry is the duration for email verification tokens (48 hours)
	EmailVerificationTokenExpiry = 48 * time.Hour
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registeredUserRequest struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Bio            string `json:"bio"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
}

type AuthHandler struct {
	UserStore              store.UserStore
	RefreshTokenStore      store.RefreshTokenStore
	PasswordResetStore     store.PasswordResetStore
	EmailVerificationStore store.EmailVerificationStore
	EmailService           *services.EmailService
	JWTService             *services.JWTService
}

func NewAuthHandler(
	userStore store.UserStore,
	refreshTokenStore store.RefreshTokenStore,
	passwordResetStore store.PasswordResetStore,
	emailVerificationStore store.EmailVerificationStore,
	emailService *services.EmailService,
	jwtService *services.JWTService,
) *AuthHandler {
	return &AuthHandler{
		UserStore:              userStore,
		RefreshTokenStore:      refreshTokenStore,
		PasswordResetStore:     passwordResetStore,
		EmailVerificationStore: emailVerificationStore,
		EmailService:           emailService,
		JWTService:             jwtService,
	}
}

// RegisterUser godoc
// @Summary Register a new user
// @Description Register a new user with the provided information
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body registeredUserRequest true "User Registration Info"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "Username or email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req registeredUserRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trim and normalize input
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Bio = strings.TrimSpace(req.Bio)
	req.ProfilePicture = strings.TrimSpace(req.ProfilePicture)

	// Required field check
	if req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, email, and password are required"})
		return
	}

	// Username checks
	if len(req.Username) < 3 || len(req.Username) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be between 3 and 20 characters"})
		return
	}
	if !utils.IsValidUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username"})
		return
	}
	if utils.IsReservedUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username not allowed"})
		return
	}

	// Email format check
	if !utils.IsValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	// Password strength check
	if len(req.Password) < 8 || !utils.ContainsNumberAndSymbol(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters with number and symbol"})
		return
	}

	// Profile picture URL check (if provided)
	if req.ProfilePicture != "" && !utils.IsValidURL(req.ProfilePicture) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile picture URL"})
		return
	}

	// Create user model
	user := &store.User{
		UserID:         uuid.New().String(),
		Username:       req.Username,
		Email:          req.Email,
		Bio:            req.Bio,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		ProfilePicture: req.ProfilePicture,
	}
	err = user.PasswordHash.SetPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set password"})
		return
	}

	// Get database connection from the store
	db := h.UserStore.(interface{ DB() *sql.DB }).DB()

	// Start a transaction for atomic operations
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
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

	// Insert user into DB within the transaction
	err = h.UserStore.CreateUserWithTransaction(user, tx)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		if strings.Contains(err.Error(), "duplicate key") {
			if strings.Contains(err.Error(), "users_username_key") {
				c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			} else if strings.Contains(err.Error(), "users_email_key") {
				c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			} else {
				c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	// Generate JWT tokens
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()
	fmt.Printf("User %s signed up from IP: %s, User-Agent: %s\n", user.Username, ipAddress, userAgent)

	// Generate tokens within the transaction
	accessToken, refreshToken, err := h.JWTService.GenerateTokenPairWithTransaction(user, ipAddress, userAgent, tx)
	if err != nil {
		log.Printf("Failed to generate token pair: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate auth tokens"})
		return
	}

	// Commit the transaction since both user creation and token generation succeeded
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Generate a verification token and send verification email
	if h.EmailVerificationStore != nil && h.EmailService != nil {
		go func() {
			// Create a verification token
			token, err := h.EmailVerificationStore.CreateVerificationToken(user.UserID, EmailVerificationTokenExpiry)
			if err != nil {
				log.Printf("Failed to create verification token for %s: %v", user.Email, err)
				return
			}

			// Send verification email
			name := user.FirstName
			if name == "" {
				name = user.Username
			}

			emailID, err := h.EmailService.SendVerificationEmail(user.Email, name, token.Token)
			if err != nil {
				log.Printf("Failed to send verification email to %s: %v", user.Email, err)
			} else {
				log.Printf("Verification email sent to %s with ID: %s", user.Email, emailID)
			}
		}()
	} else {
		// Fall back to welcome email if verification store is not available
		if h.EmailService != nil {
			go func() {
				name := user.FirstName
				if name == "" {
					name = user.Username
				}
				emailID, err := h.EmailService.SendWelcomeEmail(user.Email, name)
				if err != nil {
					log.Printf("Failed to send welcome email to %s: %v", user.Email, err)
				} else {
					log.Printf("Welcome email sent to %s with ID: %s", user.Email, emailID)
				}
			}()
		}
	}

	// Return success with tokens
	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
		"tokens": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken.Token,
		},
		"user": gin.H{
			"user_id":         user.UserID,
			"username":        user.Username,
			"email":           user.Email,
			"bio":             user.Bio,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"profile_picture": user.ProfilePicture,
			"email_verified":  user.EmailVerified,
			"created_at":      user.CreatedAt,
		},
	})
}

// LoginUser godoc
// @Summary User login
// @Description Authenticates a user and returns access and refresh tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body loginRequest true "User login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with user info and tokens"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) LoginUser(c *gin.Context) {
	var req loginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Get user by email
	user, err := h.UserStore.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Login error looking up user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Verify password
	err = user.PasswordHash.CheckPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Update last_login timestamp
	err = h.UserStore.UpdateLastLogin(user.UserID)
	if err != nil {
		log.Printf("Failed to update last_login: %v", err)
		// Continue with login process despite the error in updating last_login
	}

	// Generate JWT tokens
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	accessToken, refreshToken, err := h.JWTService.GenerateTokenPair(user, ipAddress, userAgent)
	if err != nil {
		log.Printf("Failed to generate token pair: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate auth tokens"})
		return
	}

	// No longer setting cookies as tokens will be stored in localStorage

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"tokens": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken.Token,
		},
		"user": gin.H{
			"user_id":         user.UserID,
			"username":        user.Username,
			"email":           user.Email,
			"bio":             user.Bio,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"profile_picture": user.ProfilePicture,
			"email_verified":  user.EmailVerified,
			"created_at":      user.CreatedAt,
			"last_login":      user.LastLogin,
		},
	})
}

// LogoutUser godoc
// @Summary Logout user
// @Description Ends the current user session by revoking the refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token to revoke"
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 400 {object} map[string]string "Missing refresh token"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/logout [post]
// @Security BearerAuth
func (h *AuthHandler) LogoutUser(c *gin.Context) {
	// Get refresh token from request body
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh token"})
		return
	}

	refreshTokenString := req.RefreshToken
	if refreshTokenString == "" {
		c.JSON(http.StatusOK, gin.H{"message": "no active session"})
		return
	}

	// Revoke refresh token in DB
	err := h.JWTService.RevokeRefreshToken(refreshTokenString)
	if err != nil {
		log.Printf("Failed to revoke refresh token: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// RefreshAccessToken godoc
// @Summary Refresh JWT access token
// @Description Validates refresh token and issues a new access token with token rotation
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} map[string]interface{} "New access and refresh tokens"
// @Failure 401 {object} map[string]string "Invalid or expired refresh token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/token/refresh [post]
func (h *AuthHandler) RefreshAccessToken(c *gin.Context) {
	// Get refresh token from request body
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	refreshTokenString := req.RefreshToken
	if refreshTokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	// Use the token to generate a new access token and rotate the refresh token
	newAccessToken, newRefreshToken, err := h.JWTService.RefreshAccessToken(refreshTokenString)
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Return new access token and new refresh token
	c.JSON(http.StatusOK, gin.H{
		"message": "token refreshed",
		"tokens": gin.H{
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken.Token,
		},
	})
}

// GetAuthenticatedUser godoc
// @Summary Get current authenticated user
// @Description Returns the profile of the currently authenticated user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User information"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/me [get]
func (h *AuthHandler) GetAuthenticatedUser(c *gin.Context) {
	// Get user ID from context (added by AuthMiddleware)
	// Note: The JWT auth middleware will set this from the token claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get user from database
	user, err := h.UserStore.GetUserByID(userID.(string))
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Return user info - email is included as the user is authenticated and it's their own data
	// It's useful for the client to have this information for profile display and management
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"user_id":         user.UserID,
			"username":        user.Username,
			"email":           user.Email, // Email is kept as the user is viewing their own profile
			"email_verified":  user.EmailVerified,
			"bio":             user.Bio,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"profile_picture": user.ProfilePicture,
			"created_at":      user.CreatedAt,
		},
	})
}

// These helper functions have been removed as we no longer use cookies for token storage
// The frontend will store tokens in localStorage instead
