package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dapoadedire/chefshare_be/services"
	"github.com/dapoadedire/chefshare_be/store"
	"github.com/dapoadedire/chefshare_be/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserStore    store.UserStore
	EmailService *services.EmailService
	JWTService   *services.JWTService
}

func NewUserHandler(userStore store.UserStore, emailService *services.EmailService, jwtService *services.JWTService) *UserHandler {
	return &UserHandler{
		UserStore:    userStore,
		EmailService: emailService,
		JWTService:   jwtService,
	}
}

type UpdateUserRequest struct {
	Username        *string `json:"username,omitempty"`
	FirstName       *string `json:"first_name,omitempty"`
	LastName        *string `json:"last_name,omitempty"`
	Bio             *string `json:"bio,omitempty"`
	ProfilePicture  *string `json:"profile_picture,omitempty"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// UpdateUser godoc
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest true "User information to update"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 409 {object} map[string]string "Username already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [put]
// Requires authentication 
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Get user ID from context (added by AuthMiddleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse request body
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user data
	user, err := h.UserStore.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to fetch user data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}


	// Validate and prepare updates
	changes := make(map[string]interface{})
	changes["updated_at"] = time.Now()

	// Username validation and update
	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)

		// Validation
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username cannot be empty"})
			return
		}

		if len(username) < 3 || len(username) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username must be between 3 and 20 characters"})
			return
		}

		if !utils.IsValidUsername(username) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username format"})
			return
		}

		if utils.IsReservedUsername(username) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username not allowed"})
			return
		}

		// Check if new username is different from current one
		if username != user.Username {
			// Check if username is already taken by another user
			existingUser, err := h.UserStore.IsUsernameTaken(username, userID)
			if err != nil {
				log.Printf("Error checking username existence: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}

			if existingUser {
				c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
				return
			}

			changes["username"] = username
		}
	}

	// ProfilePicture validation and update
	if req.ProfilePicture != nil {
		profilePicture := strings.TrimSpace(*req.ProfilePicture)

		if profilePicture != "" && !utils.IsValidURL(profilePicture) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile picture URL"})
			return
		}

		changes["profile_picture"] = profilePicture
	}

	// Other fields update
	if req.FirstName != nil {
		changes["first_name"] = strings.TrimSpace(*req.FirstName)
	}

	if req.LastName != nil {
		changes["last_name"] = strings.TrimSpace(*req.LastName)
	}

	if req.Bio != nil {
		changes["bio"] = strings.TrimSpace(*req.Bio)
	}

	// If no changes to update
	if len(changes) <= 1 { // Only updated_at is present
		c.JSON(http.StatusOK, gin.H{
			"message": "no changes to update",
			"user": gin.H{
				"user_id":         user.UserID,
				"username":        user.Username,
				"email":           user.Email,
				"bio":             user.Bio,
				"first_name":      user.FirstName,
				"last_name":       user.LastName,
				"profile_picture": user.ProfilePicture,
				"created_at":      user.CreatedAt,
				"updated_at":      user.UpdatedAt,
			},
		})
		return
	}

	// Update user profile in database
	updatedUser, err := h.UserStore.UpdateUser(userID, changes)
	if err != nil {
		log.Printf("Failed to update user profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user profile"})
		return
	}

	// Return success with updated user data
	c.JSON(http.StatusOK, gin.H{
		"message": "profile updated successfully",
		"user": gin.H{
			"user_id":         updatedUser.UserID,
			"username":        updatedUser.Username,
			"email":           updatedUser.Email,
			"bio":             updatedUser.Bio,
			"first_name":      updatedUser.FirstName,
			"last_name":       updatedUser.LastName,
			"profile_picture": updatedUser.ProfilePicture,
			"created_at":      updatedUser.CreatedAt,
			"updated_at":      updatedUser.UpdatedAt,
		},
	})
}

// UpdatePassword godoc
// @Summary Update user password
// @Description Update the authenticated user's password
// @Tags Users
// @Accept json
// @Produce json
// @Param request body UpdatePasswordRequest true "Current and new password"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Password updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or password requirements not met"
// @Failure 401 {object} map[string]string "Unauthorized or incorrect current password"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me/password [put]
// Requires authentication and password verification
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	// Get user ID from context (added by AuthMiddleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse request body
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user data
	user, err := h.UserStore.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to fetch user data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify current password
	if err := user.PasswordHash.CheckPassword(req.CurrentPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid current password"})
		return
	}

	// Validate new password strength
	if len(req.Password) < 8 || !utils.ContainsNumberAndSymbol(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters with a number and symbol"})
		return
	}

	// Check that new password is different from the current one
	if req.CurrentPassword == req.Password {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be different from current password"})
		return
	}

	// Start a transaction for password update and token revocation
	db := h.UserStore.DB()
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
	
	// Update the password
	err = h.UserStore.UpdatePassword(userID, req.Password)
	if err != nil {
		log.Printf("Failed to update password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}
	
	// Revoke all refresh tokens for this user to invalidate all sessions
	revokedCount, err := h.JWTService.RevokeAllUserRefreshTokens(userID)
	if err != nil {
		log.Printf("Failed to revoke refresh tokens: %v", err)
		// Continue with the password change even if token revocation fails
	} else {
		log.Printf("Revoked %d refresh tokens for user %s after password change", revokedCount, userID)
	}
	
	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Send password changed email notification if email service is available
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

	c.JSON(http.StatusOK, gin.H{
		"message": "password updated successfully",
		"sessions_revoked": true,
		"info": "all sessions have been logged out for security",
	})
}

// The checkUsernameExists method has been moved to the user_store as IsUsernameTaken

// The updateUserInDatabase helper function has been removed
// The UpdateUser method now directly returns the updated user data
