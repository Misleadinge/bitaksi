package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/bitaksi/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	clients map[string]*clientLimiter
	mu      sync.RWMutex
	config  *config.RateLimitConfig
	logger  *zap.Logger
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *config.RateLimitConfig, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientLimiter),
		config:  cfg,
		logger:  logger,
	}

	// Clean up old clients periodically
	go rl.cleanup()

	return rl
}

// Limit returns a middleware that rate limits requests
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting if disabled
		if !rl.config.Enabled {
			c.Next()
			return
		}

		// Get client identifier (IP address)
		clientIP := c.ClientIP()

		// Get or create limiter for this client
		limiter := rl.getLimiter(clientIP)

		// Check if request is allowed
		if !limiter.Allow() {
			rl.logger.Warn("rate limit exceeded", zap.String("ip", clientIP))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "too many requests, please try again later",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) getLimiter(clientIP string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	client, exists := rl.clients[clientIP]
	if !exists {
		// Create new limiter: requests per window
		limiter := rate.NewLimiter(rate.Every(rl.config.Window/time.Duration(rl.config.Requests)), rl.config.Requests)
		rl.clients[clientIP] = &clientLimiter{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	client.lastSeen = time.Now()
	return client.limiter
}

// cleanup removes old clients that haven't been seen in a while
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, client := range rl.clients {
			if time.Since(client.lastSeen) > 10*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}
