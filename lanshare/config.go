package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const EnvPrefix = "LC_"

type Config struct {
	Port         int
	DownloadPath string
	MaxFileSize  int64
	AllowedTypes []string
}

// getDefaultConfig returns a Config struct with default values
func getDefaultConfig() *Config {
	return &Config{
		Port:         5666,
		DownloadPath: "./download",
		MaxFileSize:  100 * 1024 * 1024, // 100MB
		AllowedTypes: []string{"*"},
	}
}

// getEnvAsInt reads an environment variable and converts it to int
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := GetEnv(name)
	if valueStr == "" {
		return defaultVal
	}

	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	log.Printf("Warning: Invalid value for %s: %s, using default: %d", name, valueStr, defaultVal)
	return defaultVal
}

// getEnvAsInt64 reads an environment variable and converts it to int64
func getEnvAsInt64(name string, defaultVal int64) int64 {
	valueStr := GetEnv(name)
	if valueStr == "" {
		return defaultVal
	}

	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}

	log.Printf("Warning: Invalid value for %s: %s, using default: %d", name, valueStr, defaultVal)
	return defaultVal
}

// getEnvAsString reads an environment variable as string
func getEnvAsString(name string, defaultVal string) string {
	if value := GetEnv(name); value != "" {
		return value
	}
	return defaultVal
}

// getEnvAsStringSlice reads an environment variable and splits it by comma
func getEnvAsStringSlice(name string, defaultVal []string) []string {
	valueStr := GetEnv(name)
	if valueStr == "" {
		return defaultVal
	}

	values := strings.Split(valueStr, ",")
	for i, v := range values {
		values[i] = strings.TrimSpace(v)
	}

	return values
}

// validateConfig validates configuration values
func validateConfig(config *Config) error {
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", config.Port)
	}

	if config.MaxFileSize <= 0 {
		return fmt.Errorf("invalid max file size %d: must be positive", config.MaxFileSize)
	}

	if config.DownloadPath == "" {
		return fmt.Errorf("download path cannot be empty")
	}

	return nil
}

func LoadConfig() *Config {
	log.Printf("Loading configuration from environment variables")

	// Get default config
	defaultConfig := getDefaultConfig()

	// Load configuration from environment variables
	config := &Config{
		Port:         getEnvAsInt("PORT", defaultConfig.Port),
		DownloadPath: getEnvAsString("DOWNLOAD_PATH", defaultConfig.DownloadPath),
		MaxFileSize:  getEnvAsInt64("MAX_FILE_SIZE", defaultConfig.MaxFileSize),
		AllowedTypes: getEnvAsStringSlice("ALLOWED_TYPES", defaultConfig.AllowedTypes),
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Printf("Warning: Configuration validation failed: %v", err)
		log.Printf("Using default values for invalid fields")

		// Fix invalid values with defaults
		if config.Port < 1 || config.Port > 65535 {
			config.Port = defaultConfig.Port
		}
		if config.MaxFileSize <= 0 {
			config.MaxFileSize = defaultConfig.MaxFileSize
		}
		if config.DownloadPath == "" {
			config.DownloadPath = defaultConfig.DownloadPath
		}
		if len(config.AllowedTypes) == 0 {
			config.AllowedTypes = defaultConfig.AllowedTypes
		}
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Final config - Port: %d, DownloadPath: %s, MaxFileSize: %d bytes, AllowedTypes: %d types",
		config.Port, config.DownloadPath, config.MaxFileSize, len(config.AllowedTypes))

	return config
}

func GetEnv(key string) string {
	return os.Getenv(EnvPrefix + key)
}
