package middleware

import (
	"net/http"
	"strings"

	"github.com/dapoadedire/chefshare_be/services"
	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware creates a middleware for JWT authentication
func JWTAuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		
		// Check for Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}
		
		accessToken := parts[1]
		
		// Validate the token
		claims, err := jwtService.ValidateAccessToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		
		// Set claims in context for handlers to use
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		
		// Continue processing the request
		c.Next()
	}
}




