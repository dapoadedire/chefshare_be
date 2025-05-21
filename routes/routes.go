package routes

import (
	"github.com/dapoadedire/chefshare_be/app"
	"github.com/gin-gonic/gin"
	"time"
)

func SetupRoutes(app *app.Application) *gin.Engine {
	router := gin.Default()

	// API version group
	v1 := router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("", app.UserHandler.CreateUser)

		}

		// Health check endpoint
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":   "ok",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})
	}
	return router
}