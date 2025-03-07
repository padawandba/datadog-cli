package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	config := &Config{
		Site:   "datadoghq.com",
		Output: "table",
	}

	// Try to load from config file
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".config", "dd", "config.json")
		if _, err := os.Stat(configPath); err == nil {
			file, err := os.Open(configPath)
			if err == nil {
				defer file.Close()
				if err := json.NewDecoder(file).Decode(config); err != nil {
					return nil, fmt.Errorf("failed to parse config file: %v", err)
				}
			}
		}
	}

	// Override with environment variables
	if apiKey := os.Getenv("DD_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}
	if appKey := os.Getenv("DD_APP_KEY"); appKey != "" {
		config.AppKey = appKey
	}
	if site := os.Getenv("DD_SITE"); site != "" {
		config.Site = site
	}

	return config, nil
}

// Save saves the configuration to the config file
func Save(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "ddadmin")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
