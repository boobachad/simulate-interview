package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ProviderConfig represents the provider configuration
type ProviderConfig struct {
	ActiveProvider string `json:"active_provider"`
	Gemini         struct {
		Model string `json:"model"`
	} `json:"gemini"`
	OpenRouter struct {
		Model string `json:"model"`
	} `json:"openrouter"`
}

var Config ProviderConfig

// LoadConfig loads the configuration from config.json
func LoadConfig() error {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("failed to read config.json: %w", err)
	}

	err = json.Unmarshal(file, &Config)
	if err != nil {
		return fmt.Errorf("failed to parse config.json: %w", err)
	}

	return nil
}
