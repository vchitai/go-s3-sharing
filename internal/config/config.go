package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the S3 sharing service
type Config struct {
	Server   ServerConfig
	AWS      AWSConfig
	Redis    RedisConfig
	Security SecurityConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// AWSConfig holds AWS S3 configuration
type AWSConfig struct {
	Region string
	Bucket string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr       string
	Password   string
	DB         int
	TLSEnabled bool
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	MaxAgeDays int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 120*time.Second),
		},
		AWS: AWSConfig{
			Region: getEnv("AWS_REGION", "us-east-1"),
			Bucket: getEnv("S3_BUCKET", ""),
		},
		Redis: RedisConfig{
			Addr:       getEnv("REDIS_ADDR", "localhost:6379"),
			Password:   getEnv("REDIS_PASSWORD", ""),
			DB:         getIntEnv("REDIS_DB", 0),
			TLSEnabled: getBoolEnv("REDIS_TLS_ENABLED", false),
		},
		Security: SecurityConfig{
			MaxAgeDays: getIntEnv("MAX_AGE_DAYS", 90),
		},
	}

	// Validate required fields
	if cfg.AWS.Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true"
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
