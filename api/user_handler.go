package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/dapoadedire/chefshare_be/services"
	"github.com/dapoadedire/chefshare_be/store"
	"github.com/dapoadedire/chefshare_be/utils"
	"github.com/gin-gonic/gin"
)

type registeredUserRequest struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Bio            string `json:"bio"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
}

type UserHandler struct {
	UserStore    store.UserStore
	EmailService *services.EmailService
}

func NewUserHandler(userStore store.UserStore, emailService *services.EmailService) *UserHandler {
	return &UserHandler{
		UserStore:    userStore,
		EmailService: emailService,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
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

	c.JSON(http.StatusCreated, user)
}
