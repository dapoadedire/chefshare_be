package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/dapoadedire/chefshare_be/store"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a middleware for authentication
func AuthMiddleware(sessionStore store.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the cookie
		token, err := c.Cookie("auth_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		// Get the session from the database
		session, err := sessionStore.GetSessionByToken(token)
		if err != nil {
			log.Printf("Error retrieving session: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// Check if session exists and is valid
		if session == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			return
		}

		// Check if session is expired
		if session.ExpiresAt.Before(time.Now()) {
			// Clean up expired session
			_ = sessionStore.DeleteSession(token)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		// Set user ID in context for handlers to use
		c.Set("user_id", session.UserID)
		c.Set("session_token", session.Token)

		// Continue processing the request
		c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware that authenticates if a token is present but doesn't abort if not
func OptionalAuthMiddleware(sessionStore store.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the cookie
		token, err := c.Cookie("auth_token")
		if err != nil {
			// No cookie, continue without authentication
			c.Next()
			return
		}

		// Get the session from the database
		session, err := sessionStore.GetSessionByToken(token)
		if err != nil {
			// Log error but continue
			log.Printf("Error retrieving session: %v", err)
			c.Next()
			return
		}

		// Check if session exists and is not expired
		if session != nil && session.ExpiresAt.After(time.Now()) {
			// Set user ID in context
			c.Set("user_id", session.UserID)
			c.Set("session_token", session.Token)
		}

		// Continue processing the request
		c.Next()
	}
}

// SessionCleanupMiddleware periodically cleans up expired sessions
func SessionCleanupMiddleware(sessionStore store.SessionStore, cleanupInterval time.Duration) gin.HandlerFunc {
	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			count, err := sessionStore.DeleteExpiredSessions()
			if err != nil {
				log.Printf("Error cleaning up expired sessions: %v", err)
			} else if count > 0 {
				log.Printf("Cleaned up %d expired sessions", count)
			}
		}
	}()

	// Return a pass-through middleware
	return func(c *gin.Context) {
		c.Next()
	}
}
