package middleware

import (
	"net/http"
	"strings"

	"github.com/bitaksi/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// APIKeyAuth returns a middleware that validates API keys
func APIKeyAuth(cfg *config.Config, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip API key check if disabled
		if !cfg.APIKey.Enabled {
			c.Next()
			return
		}

		// Get API key from header (support multiple header formats)
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try Authorization header with "ApiKey" prefix
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "apikey" {
					apiKey = parts[1]
				}
			}
		}

		if apiKey == "" {
			logger.Debug("API key missing")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "API key is required",
				},
			})
			c.Abort()
			return
		}

		// Validate API key
		if !isValidAPIKey(apiKey, cfg.APIKey.Keys) {
			logger.Warn("invalid API key attempted", zap.String("key_prefix", maskAPIKey(apiKey)))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid API key",
				},
			})
			c.Abort()
			return
		}

		// Set API key in context for logging/auditing
		c.Set("api_key", maskAPIKey(apiKey))
		c.Next()
	}
}

// isValidAPIKey checks if the provided API key is valid
func isValidAPIKey(key string, validKeys []string) bool {
	if len(validKeys) == 0 {
		return false
	}

	for _, validKey := range validKeys {
		if key == validKey {
			return true
		}
	}

	return false
}

// maskAPIKey masks the API key for logging (shows first 8 and last 4 characters)
func maskAPIKey(key string) string {
	if len(key) <= 12 {
		return "****"
	}
	return key[:8] + "****" + key[len(key)-4:]
}
