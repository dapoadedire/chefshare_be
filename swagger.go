package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dapoadedire/chefshare_be/docs"
)

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