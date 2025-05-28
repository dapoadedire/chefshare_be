package middleware

import (
	"log"
	"time"

	"github.com/dapoadedire/chefshare_be/store"
	"github.com/gin-gonic/gin"
)

// PasswordResetCleanupMiddleware periodically cleans up expired password reset tokens
func PasswordResetCleanupMiddleware(passwordResetStore store.PasswordResetStore, cleanupInterval time.Duration) gin.HandlerFunc {
	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			count, err := passwordResetStore.DeleteExpiredTokens()
			if err != nil {
				log.Printf("Error cleaning up expired password reset tokens: %v", err)
			} else if count > 0 {
				log.Printf("Cleaned up %d expired password reset tokens", count)
			}
		}
	}()

	// Return a pass-through middleware
	return func(c *gin.Context) {
		c.Next()
	}
}
