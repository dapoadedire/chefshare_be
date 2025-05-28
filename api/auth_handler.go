package api

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dapoadedire/chefshare_be/services"
	"github.com/dapoadedire/chefshare_be/store"
	"github.com/dapoadedire/chefshare_be/utils"
	"github.com/gin-gonic/gin"
)

const (
	// DefaultSessionDuration is the default duration for sessions (7 days)
	DefaultSessionDuration = 7 * 24 * time.Hour
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
	UserStore    store.UserStore
	SessionStore store.SessionStore
	EmailService *services.EmailService
}

func NewAuthHandler(userStore store.UserStore, sessionStore store.SessionStore, emailService *services.EmailService) *AuthHandler {
	return &AuthHandler{
		UserStore:    userStore,
		SessionStore: sessionStore,
		EmailService: emailService,
	}
}

// SignUp creates a new user and establishes a session
func (h *AuthHandler) SignUp(c *gin.Context) {
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

	// Insert into DB
	err = h.UserStore.CreateUser(user)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	// Create session
	session, err := h.SessionStore.CreateSession(int64(user.ID), DefaultSessionDuration)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Set cookie
	setCookieForSession(c, session)

	// Send welcome email async
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

	// Return success without exposing the session token in the response body
	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
		"user": gin.H{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"bio":             user.Bio,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"profile_picture": user.ProfilePicture,
			"created_at":      user.CreatedAt,
		},
	})
}

// Login validates credentials and establishes a session
func (h *AuthHandler) Login(c *gin.Context) {
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

	// Create session
	session, err := h.SessionStore.CreateSession(int64(user.ID), DefaultSessionDuration)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Set cookie
	setCookieForSession(c, session)

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"user": gin.H{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"bio":             user.Bio,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"profile_picture": user.ProfilePicture,
			"created_at":      user.CreatedAt,
		},
	})
}

// Logout terminates the current session
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from cookie
	token, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "no active session"})
		return
	}

	// Delete session from DB
	err = h.SessionStore.DeleteSession(token)
	if err != nil {
		log.Printf("Failed to delete session: %v", err)
		// Continue to clear cookie anyway
	}

	// Clear cookie
	domain := getDomainFromEnv()
	c.SetCookie("auth_token", "", -1, "/", domain, true, true)

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (added by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get user from database
	user, err := h.UserStore.GetUserByID(userID.(int64))
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Return user info
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"bio":             user.Bio,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"profile_picture": user.ProfilePicture,
			"created_at":      user.CreatedAt,
		},
	})
}

// Helper to set a cookie for a session
func setCookieForSession(c *gin.Context, session *store.Session) {
	domain := getDomainFromEnv()
	maxAge := int(DefaultSessionDuration.Seconds())

	c.SetCookie(
		"auth_token",  // name
		session.Token, // value
		maxAge,        // max age in seconds
		"/",           // path
		domain,        // domain
		true,          // secure (HTTPS only)
		true,          // HTTP only (not accessible by JavaScript)
	)
}

// Helper to get the domain for cookies from environment variable
func getDomainFromEnv() string {
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		return "localhost" // Default for local development
	}
	return domain
}
