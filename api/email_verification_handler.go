package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dapoadedire/chefshare_be/middleware"
	"github.com/dapoadedire/chefshare_be/utils"
	"github.com/gin-gonic/gin"
)

// Using EmailVerificationTokenExpiry from auth_handler.go

type verifyEmailRequest struct {
	Token string `json:"token"`
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verifies a user's email address using the token sent in the verification email
// @Tags Email Verification
// @Accept json
// @Produce json
// @Param request body verifyEmailRequest true "Verification token"
// @Success 200 {object} map[string]string "Email verified successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Token not found"
// @Failure 410 {object} map[string]string "Token expired"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/verify-email/confirm [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Trim the token
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	// Get verification token
	token, err := h.EmailVerificationStore.GetVerificationTokenByToken(req.Token)
	if err != nil {
		log.Printf("Error retrieving email verification token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if token == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid or expired verification token"})
		return
	}

	// Check if token has expired
	if token.ExpiresAt.Before(time.Now()) {
		// Delete the expired token
		err = h.EmailVerificationStore.DeleteToken(token.ID)
		if err != nil {
			log.Printf("Error deleting expired token: %v", err)
		}
		c.JSON(http.StatusGone, gin.H{"error": "verification link has expired, please request a new one"})
		return
	}

	// Get user by ID
	user, err := h.UserStore.GetUserByID(token.UserID)
	if err != nil || user == nil {
		log.Printf("Error retrieving user for email verification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Check if email is already verified
	if user.EmailVerified {
		// Delete the token as it's no longer needed
		err = h.EmailVerificationStore.DeleteToken(token.ID)
		if err != nil {
			log.Printf("Error deleting token after verification: %v", err)
		}
		c.JSON(http.StatusOK, gin.H{"message": "email is already verified"})
		return
	}

	// Mark the email as verified
	err = h.UserStore.SetEmailVerified(token.UserID, true)
	if err != nil {
		log.Printf("Error marking email as verified: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	// Delete the token as it's no longer needed
	err = h.EmailVerificationStore.DeleteToken(token.ID)
	if err != nil {
		log.Printf("Error deleting token after verification: %v", err)
		// Continue despite this error since the email was verified
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

// ResendVerificationEmail godoc
// @Summary Resend verification email
// @Description Sends a new verification email to the user
// @Tags Email Verification
// @Accept json
// @Produce json
// @Param request body resendVerificationRequest true "Email address"
// @Success 200 {object} map[string]string "Verification email sent"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 429 {object} map[string]string "Rate limit exceeded"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/verify-email/resend [post]
func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	var req resendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	if !utils.IsValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	// Apply email-based rate limiting
	if !middleware.TrackEmailRateLimiting(req.Email) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": "too many verification attempts, please try again later",
		})
		return
	}

	// Get user by email
	user, err := h.UserStore.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Error looking up user for verification email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		// For security reasons, don't reveal whether the email exists or not
		c.JSON(http.StatusOK, gin.H{"message": "if your email is registered and not verified, a verification email will be sent"})
		return
	}

	// If email is already verified, no need to send a new verification email
	if user.EmailVerified {
		c.JSON(http.StatusOK, gin.H{"message": "email is already verified"})
		return
	}

	// Generate a new verification token
	token, err := h.EmailVerificationStore.CreateVerificationToken(user.UserID, EmailVerificationTokenExpiry)
	if err != nil {
		log.Printf("Error creating verification token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create verification token"})
		return
	}

	// Send verification email
	if h.EmailService != nil {
		name := user.FirstName
		if name == "" {
			name = user.Username
		}

		go func() {
			emailID, err := h.EmailService.SendVerificationEmail(user.Email, name, token.Token)
			if err != nil {
				log.Printf("Failed to send verification email to %s: %v", user.Email, err)
			} else {
				log.Printf("Verification email sent to %s with ID: %s", user.Email, emailID)
			}
		}()
	} else {
		log.Printf("Email service not available, verification token for %s is: %s", user.Email, token.Token)
	}

	c.JSON(http.StatusOK, gin.H{"message": "if your email is registered and not verified, a verification email will be sent"})
}

// RequestVerificationEmail godoc
// @Summary Request verification email (authenticated)
// @Description Sends a new verification email to the authenticated user
// @Tags Email Verification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Verification email sent"
// @Failure 400 {object} map[string]string "Email already verified"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 429 {object} map[string]string "Rate limit exceeded"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/verify-email/request [post]
func (h *AuthHandler) RequestVerificationEmail(c *gin.Context) {
	// Get user ID from context (added by AuthMiddleware)
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

	// If email is already verified, no need to send a new verification email
	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is already verified"})
		return
	}

	// Apply email-based rate limiting
	if !middleware.TrackEmailRateLimiting(user.Email) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": "too many verification attempts, please try again later",
		})
		return
	}

	// Generate a new verification token
	token, err := h.EmailVerificationStore.CreateVerificationToken(user.UserID, EmailVerificationTokenExpiry)
	if err != nil {
		log.Printf("Error creating verification token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create verification token"})
		return
	}

	// Send verification email
	if h.EmailService != nil {
		name := user.FirstName
		if name == "" {
			name = user.Username
		}

		go func() {
			emailID, err := h.EmailService.SendVerificationEmail(user.Email, name, token.Token)
			if err != nil {
				log.Printf("Failed to send verification email to %s: %v", user.Email, err)
			} else {
				log.Printf("Verification email sent to %s with ID: %s", user.Email, emailID)
			}
		}()
	} else {
		log.Printf("Email service not available, verification token for %s is: %s", user.Email, token.Token)
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification email sent"})
}
