package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dapoadedire/chefshare_be/utils"
	"github.com/gin-gonic/gin"
)

const (
	// OTPExpiry duration for OTP tokens (15 minutes)
	OTPExpiry = 15 * time.Minute
)

type requestOTPRequest struct {
	Email string `json:"email"`
}

type verifyOTPRequest struct {
	Email    string `json:"email"`
	OTP      string `json:"otp"`
	Password string `json:"password"`
}

type resendOTPRequest struct {
	Email string `json:"email"`
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Initiates the password reset process by sending an OTP to the user's email
// @Tags Password Reset
// @Accept json
// @Produce json
// @Param request body requestOTPRequest true "Email for reset"
// @Success 200 {object} map[string]string "OTP sent to email"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/forgot-password [post]
// RequestPasswordReset initiates the password reset process by sending an OTP
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req requestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validate email format
	if !utils.IsValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	// Get user by email
	user, err := h.UserStore.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Error looking up user for password reset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// If user not found, we still return success to prevent email enumeration
	// But we don't actually send an email
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "if your email is registered, we've sent a password reset code",
		})
		return
	}

	// Generate and store OTP
	token, err := h.PasswordResetStore.CreatePasswordResetToken(int64(user.ID), OTPExpiry)
	if err != nil {
		log.Printf("Error creating password reset token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reset token"})
		return
	}

	// Send OTP via email
	if h.EmailService != nil {
		name := user.FirstName
		if name == "" {
			name = user.Username
		}

		go func() {
			emailID, err := h.EmailService.SendPasswordResetEmail(user.Email, name, token.Token)
			if err != nil {
				log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
			} else {
				log.Printf("Password reset email sent to %s with ID: %s", user.Email, emailID)
			}
		}()
	} else {
		log.Printf("Email service not available, OTP for %s is: %s", user.Email, token.Token)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "if your email is registered, we've sent a password reset code",
	})
}

// VerifyOTPAndResetPassword godoc
// @Summary Verify OTP and reset password
// @Description Verifies the OTP sent to user's email and resets the password
// @Tags Password Reset
// @Accept json
// @Produce json
// @Param request body verifyOTPRequest true "OTP verification and new password"
// @Success 200 {object} map[string]string "Password reset successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid OTP"
// @Failure 404 {object} map[string]string "User not found or OTP expired"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/reset-password [post]
// VerifyOTPAndResetPassword verifies the OTP and sets a new password
func (h *AuthHandler) VerifyOTPAndResetPassword(c *gin.Context) {
	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email and trim inputs
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.OTP = strings.TrimSpace(req.OTP)

	// Validate email format
	if !utils.IsValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	// Validate OTP format (6 digits)
	if len(req.OTP) != 6 || !utils.IsNumeric(req.OTP) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid OTP format"})
		return
	}

	// Validate password strength
	if len(req.Password) < 8 || !utils.ContainsNumberAndSymbol(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters with a number and symbol"})
		return
	}

	// Get user by email
	user, err := h.UserStore.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Error looking up user for password reset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get token by OTP value
	token, err := h.PasswordResetStore.GetPasswordResetTokenByToken(req.OTP)
	if err != nil {
		log.Printf("Error retrieving password reset token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Verify token exists, matches user, is not used, and is not expired
	if token == nil || token.UserID != int64(user.ID) || token.Used || token.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired OTP"})
		return
	}

	// Update user's password
	err = h.UserStore.UpdatePassword(int64(user.ID), req.Password)
	if err != nil {
		log.Printf("Error updating password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Mark token as used
	err = h.PasswordResetStore.MarkTokenAsUsed(token.ID)
	if err != nil {
		log.Printf("Error marking token as used: %v", err)
		// Continue processing as the password has already been changed
	}

	// Invalidate all sessions for this user to force re-login with new password
	_, err = h.SessionStore.DeleteUserSessions(int64(user.ID))
	if err != nil {
		log.Printf("Error deleting user sessions: %v", err)
		// Continue as this is not critical
	}

	// Send confirmation email
	if h.EmailService != nil {
		name := user.FirstName
		if name == "" {
			name = user.Username
		}

		go func() {
			_, err := h.EmailService.SendPasswordChangedEmail(user.Email, name)
			if err != nil {
				log.Printf("Failed to send password changed email to %s: %v", user.Email, err)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}

// ResendOTP godoc
// @Summary Resend OTP
// @Description Resends the OTP to the user's email for password reset
// @Tags Password Reset
// @Accept json
// @Produce json
// @Param request body resendOTPRequest true "Email for OTP resend"
// @Success 200 {object} map[string]string "OTP resent successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/resend-otp [post]
// ResendOTP resends the OTP to the user's email
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req resendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validate email format
	if !utils.IsValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	// Get user by email
	user, err := h.UserStore.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Error looking up user for password reset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// If user not found, we still return success to prevent email enumeration
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "if your email is registered, we've sent a new password reset code",
		})
		return
	}

	// Generate and store a new OTP (this will invalidate any existing ones)
	token, err := h.PasswordResetStore.CreatePasswordResetToken(int64(user.ID), OTPExpiry)
	if err != nil {
		log.Printf("Error creating password reset token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reset token"})
		return
	}

	// Send OTP via email
	if h.EmailService != nil {
		name := user.FirstName
		if name == "" {
			name = user.Username
		}

		go func() {
			emailID, err := h.EmailService.SendPasswordResetEmail(user.Email, name, token.Token)
			if err != nil {
				log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
			} else {
				log.Printf("Password reset email sent to %s with ID: %s", user.Email, emailID)
			}
		}()
	} else {
		log.Printf("Email service not available, OTP for %s is: %s", user.Email, token.Token)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "if your email is registered, we've sent a new password reset code",
	})
}
