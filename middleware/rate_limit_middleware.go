package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Simple in-memory rate limiter
type RateLimiter struct {
	limits       map[string][]time.Time
	mu           sync.Mutex
	windowLength time.Duration
	maxRequests  int
}

// NewRateLimiter creates a new rate limiter with the given window length and max requests
func NewRateLimiter(windowLength time.Duration, maxRequests int) *RateLimiter {
	// Start a background goroutine to periodically clean up old entries
	limiter := &RateLimiter{
		limits:       make(map[string][]time.Time),
		windowLength: windowLength,
		maxRequests:  maxRequests,
	}

	// Start cleanup routine
	go func() {
		for {
			time.Sleep(windowLength)
			limiter.cleanup()
		}
	}()

	return limiter
}

// cleanup removes expired entries from the limits map
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, times := range rl.limits {
		var validTimes []time.Time
		for _, t := range times {
			if now.Sub(t) <= rl.windowLength {
				validTimes = append(validTimes, t)
			}
		}

		if len(validTimes) == 0 {
			delete(rl.limits, key)
		} else {
			rl.limits[key] = validTimes
		}
	}
}

// Allow checks if a new request is allowed and updates the rate limiter
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Initialize if this is the first request for this key
	if _, exists := rl.limits[key]; !exists {
		rl.limits[key] = []time.Time{now}
		return true
	}

	// Filter out timestamps outside the window
	var validTimes []time.Time
	for _, t := range rl.limits[key] {
		if now.Sub(t) <= rl.windowLength {
			validTimes = append(validTimes, t)
		}
	}

	// Check if we're over the limit
	if len(validTimes) >= rl.maxRequests {
		rl.limits[key] = validTimes
		return false
	}

	// Add the current request time and allow
	rl.limits[key] = append(validTimes, now)
	return true
}

// Global rate limiters for password reset endpoints
var (
	// IP-based limiter: 5 requests per IP per 10 minutes
	ipLimiter = NewRateLimiter(10*time.Minute, 5)

	// Email-based limiter: 3 requests per email per hour
	emailLimiter = NewRateLimiter(60*time.Minute, 3)
)

// PasswordResetRateLimitMiddleware provides IP-based rate limiting for password reset endpoints
// Since we can't reliably extract the email without consuming the request body, we only use IP-based limiting
func PasswordResetRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Apply IP-based rate limiting
		if !ipLimiter.Allow(clientIP) {
			// Return a 429 Too Many Requests response
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many password reset attempts, please try again later",
			})
			c.Abort()
			return
		}

		// Continue processing
		c.Next()
	}
}

// TrackEmailRateLimiting tracks email-based rate limiting
// This should be called explicitly from within handler functions after parsing the email
func TrackEmailRateLimiting(email string) bool {
	return emailLimiter.Allow(email)
}
