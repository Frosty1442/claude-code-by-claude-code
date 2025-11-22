package config

import (
	"fmt"
	"os"
)

// Config holds application configuration
type Config struct {
	// API Configuration
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64

	// Working Directory
	WorkDir string

	// System Prompt
	SystemPrompt string
}

// LoadConfig loads configuration from environment, YAML file, and defaults
func LoadConfig() (*Config, error) {
	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load from environment first
	envCfg := &Config{
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		BaseURL:     os.Getenv("OPENAI_BASE_URL"),
		Model:       os.Getenv("MODEL"),
		MaxTokens:   4096,
		Temperature: 0.7,
		WorkDir:     workDir,
	}

	// Try to load YAML config
	yamlCfg, _ := LoadYAMLConfig(workDir)

	// Merge configs (env takes precedence)
	cfg := MergeWithEnv(yamlCfg, envCfg)

	// Set defaults if still not set
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}

	if cfg.Model == "" {
		cfg.Model = "gpt-4"
	}

	// Validate required config
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable or config.yaml api.key is required")
	}

	// Set system prompt
	cfg.SystemPrompt = GetSystemPrompt()

	return cfg, nil
}
