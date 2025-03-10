package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Common errors
var (
	ErrConfigNotFound     = errors.New("configuration file not found")
	ErrConfigParseFailure = errors.New("failed to parse configuration file")
	ErrConfigWriteFailure = errors.New("failed to write configuration file")
)

// Config holds the application configuration
type Config struct {
	APIKey string `json:"dd_api_key"`
	AppKey string `json:"dd_app_key"`
	Site   string `json:"dd_site"`
	Output string `json:"output"`
}

// Load loads configuration from config file and environment variables
func Load() (*Config, error) {
	// Initialize with default values
	config := &Config{
		Site:   "datadoghq.com", // The 'api.' prefix will be added by the client
		Output: "table",
	}

	// Try to load from config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("Could not determine user home directory", "error", err)
	} else {
		configPath := filepath.Join(homeDir, ".config", "dd", "config.json")
		loaded, err := loadFromFile(configPath, config)
		if err != nil && !errors.Is(err, ErrConfigNotFound) {
			// Only return error if it's not just a missing config file
			return nil, fmt.Errorf("config error: %w", err)
		}
		if loaded {
			slog.Debug("Loaded configuration from file", "path", configPath)
		}
	}

	// Override with environment variables
	envLoaded := false
	
	if apiKey := os.Getenv("DD_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		envLoaded = true
	}
	if appKey := os.Getenv("DD_APP_KEY"); appKey != "" {
		config.AppKey = appKey
		envLoaded = true
	}
	if site := os.Getenv("DD_SITE"); site != "" {
		config.Site = site
		envLoaded = true
	}
	
	if envLoaded {
		slog.Debug("Applied environment variable configuration")
	}

	return config, nil
}

// loadFromFile loads configuration from a file
func loadFromFile(path string, config *Config) (bool, error) {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, ErrConfigNotFound
		}
		return false, fmt.Errorf("error checking config file: %w", err)
	}
	
	// Open and read file
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()
	
	// Decode JSON
	if err := json.NewDecoder(file).Decode(config); err != nil {
		return false, fmt.Errorf("%w: %v", ErrConfigParseFailure, err)
	}
	
	return true, nil
}

// Save saves the configuration to the config file
func Save(config *Config) error {
	if config == nil {
		return errors.New("cannot save nil configuration")
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "dd")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	
	// Create file with secure permissions
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConfigWriteFailure, err)
	}
	defer file.Close()

	// Write with pretty formatting
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigWriteFailure, err)
	}

	slog.Info("Configuration saved", "path", configPath)
	return nil
}

// Validate checks if the configuration has all required fields
func Validate(config *Config) error {
	if config == nil {
		return errors.New("configuration is nil")
	}
	
	var missingFields []string
	
	if config.APIKey == "" {
		missingFields = append(missingFields, "API key")
	}
	
	if config.AppKey == "" {
		missingFields = append(missingFields, "Application key")
	}
	
	if len(missingFields) > 0 {
		return fmt.Errorf("missing required configuration: %v", missingFields)
	}
	
	return nil
}
