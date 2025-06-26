package routes

import (
	"time"

	"github.com/dapoadedire/chefshare_be/app"
	"github.com/dapoadedire/chefshare_be/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, app *app.Application) *gin.Engine {
	// Setup password reset token cleanup middleware (run every hour)
	router.Use(middleware.PasswordResetCleanupMiddleware(app.PasswordResetStore, 1*time.Hour))
	
	// Setup refresh token cleanup (run every 12 hours)
	router.Use(func(c *gin.Context) {
		go func() {
			ticker := time.NewTicker(12 * time.Hour)
			defer ticker.Stop()
	
			for range ticker.C {
				count, err := app.RefreshTokenStore.DeleteExpiredRefreshTokens()
				if err != nil {
					// Log the error but continue
					c.Error(err)
				} else if count > 0 {
					// Log the number of tokens deleted
					// This is just for information purposes
				}
			}
		}()
		c.Next()
	})

	// Welcome endpoint
	// @Summary Welcome endpoint
	// @Description Returns a welcome message with API version
	// @Tags Welcome
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Welcome message"
	// @Router / [get]
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to ChefShare API",
			"version": "1.0.0",
		})
	})

	// API version group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint
		// @Summary Health check endpoint
		// @Description Returns the API's health status
		// @Tags Health
		// @Produce json
		// @Success 200 {object} map[string]interface{} "API is healthy"
		// @Router /health [get]
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})

		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", app.AuthHandler.SignUp)
			auth.POST("/login", app.AuthHandler.Login)
			
			// New refresh token endpoint
			auth.POST("/refresh", app.AuthHandler.RefreshToken)
			
			// Password reset flow
			auth.POST("/forgot-password", app.AuthHandler.RequestPasswordReset)
			auth.POST("/reset-password", app.AuthHandler.VerifyOTPAndResetPassword)
			auth.POST("/resend-otp", app.AuthHandler.ResendOTP)

			// Protected routes that require authentication (JWT)
			authRequired := auth.Group("")
			authRequired.Use(middleware.JWTAuthMiddleware(app.JWTService))
			{
				authRequired.GET("/me", app.AuthHandler.GetCurrentUser)
				authRequired.POST("/logout", app.AuthHandler.Logout) // Now requires auth
			}
			

		}

		// Protected routes that require JWT authentication
		protected := v1.Group("")
		protected.Use(middleware.JWTAuthMiddleware(app.JWTService))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.PUT("/update", app.UserHandler.UpdateUser)
				users.PUT("/update_password", app.UserHandler.UpdatePassword)
			}
		}
		

	}

	return router
}
