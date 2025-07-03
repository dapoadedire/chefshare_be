package middleware

import (
	"log"
	"time"

	"github.com/dapoadedire/chefshare_be/store"
	"github.com/gin-gonic/gin"
)

// TokenBlacklistCleanupMiddleware periodically cleans up expired blacklisted tokens
func TokenBlacklistCleanupMiddleware(tokenBlacklistStore store.TokenBlacklistStore, cleanupInterval time.Duration) gin.HandlerFunc {
	ticker := time.NewTicker(cleanupInterval)

	go func() {
		for range ticker.C {
			log.Println("Running scheduled cleanup of blacklisted tokens...")
			count, err := tokenBlacklistStore.CleanupExpiredTokens()
			if err != nil {
				log.Printf("Error cleaning up blacklisted tokens: %v", err)
			} else {
				log.Printf("Cleaned up %d expired blacklisted tokens", count)
			}
		}
	}()

	// Return a pass-through middleware
	return func(c *gin.Context) {
		c.Next()
	}
}
