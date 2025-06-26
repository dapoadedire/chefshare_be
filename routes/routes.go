package routes

import (
	"time"

	"github.com/dapoadedire/chefshare_be/app"
	"github.com/dapoadedire/chefshare_be/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, app *app.Application) *gin.Engine {
	// Middleware for periodic cleanups
	router.Use(middleware.PasswordResetCleanupMiddleware(app.PasswordResetStore, 1*time.Hour))
	setupRefreshTokenCleanup(router, app)

	// Root welcome route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to ChefShare API",
			"version": "1.0.0",
		})
	})

	// Versioned API routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})

		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", app.AuthHandler.RegisterUser)
			auth.POST("/login", app.AuthHandler.LoginUser)
			auth.POST("/token/refresh", app.AuthHandler.RefreshAccessToken)

			// Password reset flow
			password := auth.Group("/password/reset")
			{
				password.POST("/request", app.AuthHandler.RequestPasswordReset)
				password.POST("/confirm", app.AuthHandler.VerifyOTPAndResetPassword)
				password.POST("/resend", app.AuthHandler.ResendOTP)
			}
		}

		// Protected auth routes
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.JWTAuthMiddleware(app.JWTService))
		{
			authProtected.GET("/me", app.AuthHandler.GetAuthenticatedUser)
			authProtected.POST("/logout", app.AuthHandler.LogoutUser)
		}

		// Protected user profile routes
		users := v1.Group("/users")
		users.Use(middleware.JWTAuthMiddleware(app.JWTService))
		{
			users.PUT("/me", app.UserHandler.UpdateUser)
			users.PUT("/me/password", app.UserHandler.UpdatePassword)
		}
	}

	return router
}

// run cleanup every 12 hours in background
func setupRefreshTokenCleanup(router *gin.Engine, app *app.Application) {
	router.Use(func(c *gin.Context) {
		go func() {
			ticker := time.NewTicker(12 * time.Hour)
			defer ticker.Stop()

			for range ticker.C {
				if count, err := app.RefreshTokenStore.DeleteExpiredRefreshTokens(); err != nil {
					c.Error(err)
				} else if count > 0 {
					// Optional: log cleanup count
				}
			}
		}()
		c.Next()
	})
}
