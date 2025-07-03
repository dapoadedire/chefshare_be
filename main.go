package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dapoadedire/chefshare_be/app"
	"github.com/dapoadedire/chefshare_be/docs" // Import swagger docs
	"github.com/dapoadedire/chefshare_be/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title ChefShare API
// @version 1.0
// @description ChefShare API Documentation
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the access token.
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Set Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)
	
	// Set up Swagger host dynamically based on environment
	setupSwaggerInfo()

	// Create router
	router := gin.Default()

	// Set up middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration using gin-contrib/cors
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://chefshare-2025.vercel.app"}, // Frontend origin with fallbacks
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false, // No longer needed as we don't use cookies
		MaxAge:           12 * time.Hour,
	}))

	// Initialize application
	application, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.DB.Close()

	// Set up routes
	router = routes.SetupRoutes(router, application)

	// Set up Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.DocExpansion("list"),
		ginSwagger.DeepLinking(true)))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	// Start server
	log.Printf("Starting server on port %s...\n", port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Println("Server stopped gracefully")
	// Close the database connection

}

// setupSwaggerInfo configures swagger info dynamically based on environment
// Call this function at the beginning of your main function, before initializing the router
func setupSwaggerInfo() {
	// Check if running in production
	isProd := false
	if ginMode := os.Getenv("GIN_MODE"); ginMode == "release" {
		isProd = true
	}
	
	// Alternatively, check for deployment provider-specific environment variables
	if os.Getenv("RENDER") != "" || strings.HasSuffix(os.Getenv("HOSTNAME"), ".render.com") {
		isProd = true
	}
	
	// Set the appropriate host and scheme based on environment
	if isProd {
		// Production settings
		docs.SwaggerInfo.Host = "chefshare-be.onrender.com"
		docs.SwaggerInfo.Schemes = []string{"https"}
		fmt.Println("Swagger configured for production environment")
	} else {
		// Development settings
		docs.SwaggerInfo.Host = "localhost:8080"
		docs.SwaggerInfo.Schemes = []string{"http"}
		fmt.Println("Swagger configured for development environment")
	}
}
