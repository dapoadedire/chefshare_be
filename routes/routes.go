package routes

import (
	"time"

	"github.com/dapoadedire/chefshare_be/app"
	"github.com/dapoadedire/chefshare_be/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, app *app.Application) *gin.Engine {
	// Setup session cleanup middleware (run every 6 hours)
	router.Use(middleware.SessionCleanupMiddleware(app.SessionStore, 6*time.Hour))

	// API version group
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", app.AuthHandler.SignUp)
			auth.POST("/login", app.AuthHandler.Login)
			auth.POST("/logout", app.AuthHandler.Logout)

			// Protected route that requires authentication
			authRequired := auth.Group("")
			authRequired.Use(middleware.AuthMiddleware(app.SessionStore))
			{
				authRequired.GET("/me", app.AuthHandler.GetCurrentUser)
			}
		}

		// Protected routes that require authentication
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(app.SessionStore))
		{
			// Add your protected routes here
			// Example: protected.GET("/profile", app.UserHandler.GetProfile)
		}

		// Health check endpoint
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})
	}
	return router
}
