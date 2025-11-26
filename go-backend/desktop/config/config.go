package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the configuration for the Go backend server
type Config struct {
	// Server configuration
	Host string
	Port int

	// Logging configuration
	LogLevel  string
	LogFormat string

	// Health check configuration
	HealthCheckInterval int
}

// NewConfig creates a new Config instance with default values and environment overrides
func NewConfig() *Config {
	config := &Config{
		Host:                "localhost",
		Port:                0, // 0 means random port will be assigned
		LogLevel:            "info",
		LogFormat:           "json",
		HealthCheckInterval: 30,
	}

	// Override with environment variables if present
	if host := os.Getenv("ANY_SYNC_HOST"); host != "" {
		config.Host = host
	}

	if portStr := os.Getenv("ANY_SYNC_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}

	if logLevel := os.Getenv("ANY_SYNC_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if logFormat := os.Getenv("ANY_SYNC_LOG_FORMAT"); logFormat != "" {
		config.LogFormat = logFormat
	}

	if intervalStr := os.Getenv("ANY_SYNC_HEALTH_CHECK_INTERVAL"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil {
			config.HealthCheckInterval = interval
		}
	}

	return config
}

// GetAddress returns the full server address
func (c *Config) GetAddress() string {
	if c.Port == 0 {
		return fmt.Sprintf("%s:0", c.Host) // :0 means random port
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
