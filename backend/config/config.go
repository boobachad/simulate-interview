package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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
	ProblemGenerationStrategy string
}

var Config ProviderConfig

var validStrategies = map[string]bool{
	"rotate":  true,
	"combine": true,
	"mix":     true,
}

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

	// Load problem generation strategy from env, default to "mix"
	strategy := os.Getenv("PROBLEM_GENERATION_STRATEGY")
	if strategy == "" {
		strategy = "mix"
	}
	
	// Normalize to lowercase for case-insensitive validation
	strategy = strings.ToLower(strategy)
	
	if !validStrategies[strategy] {
		log.Printf("Invalid PROBLEM_GENERATION_STRATEGY '%s' (normalized), defaulting to 'mix'", strategy)
		strategy = "mix"
	}
	
	Config.ProblemGenerationStrategy = strategy

	return nil
}
