package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the gateway
type Config struct {
	Server        ServerConfig
	DriverService DriverServiceConfig
	Logging       LoggingConfig
	JWT           JWTConfig
	RateLimit     RateLimitConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DriverServiceConfig holds driver service configuration
type DriverServiceConfig struct {
	BaseURL string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
	Enabled    bool
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled  bool
	Requests int
	Window   time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	readTimeout, _ := strconv.Atoi(getEnv("READ_TIMEOUT_SEC", "30"))
	writeTimeout, _ := strconv.Atoi(getEnv("WRITE_TIMEOUT_SEC", "30"))
	jwtExpiration, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	rateLimitRequests, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	rateLimitWindow, _ := strconv.Atoi(getEnv("RATE_LIMIT_WINDOW_SEC", "60"))
	jwtEnabled := getEnv("JWT_ENABLED", "true") == "true"
	rateLimitEnabled := getEnv("RATE_LIMIT_ENABLED", "true") == "true"

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  time.Duration(readTimeout) * time.Second,
			WriteTimeout: time.Duration(writeTimeout) * time.Second,
		},
		DriverService: DriverServiceConfig{
			BaseURL: getEnv("DRIVER_SERVICE_URL", "http://driver-service:8081"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration: time.Duration(jwtExpiration) * time.Hour,
			Enabled:    jwtEnabled,
		},
		RateLimit: RateLimitConfig{
			Enabled:  rateLimitEnabled,
			Requests: rateLimitRequests,
			Window:   time.Duration(rateLimitWindow) * time.Second,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
