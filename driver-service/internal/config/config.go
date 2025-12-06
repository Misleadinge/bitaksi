package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the driver service
type Config struct {
	Server  ServerConfig
	MongoDB MongoDBConfig
	Logging LoggingConfig
	JWT     JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI      string
	Database string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// Load loads configuration from environment variables
func Load() *Config {
	readTimeout, _ := strconv.Atoi(getEnv("READ_TIMEOUT_SEC", "30"))
	writeTimeout, _ := strconv.Atoi(getEnv("WRITE_TIMEOUT_SEC", "30"))

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8081"),
			ReadTimeout:  time.Duration(readTimeout) * time.Second,
			WriteTimeout: time.Duration(writeTimeout) * time.Second,
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "taxihub"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
