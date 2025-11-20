package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAMLConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Create config.yaml
	configPath := filepath.Join(claudeDir, "config.yaml")
	configContent := `
api:
  provider: openai
  key: test-api-key
  base_url: https://api.test.com/v1
  model: gpt-4
  max_tokens: 8192
  temperature: 0.8
`
	os.WriteFile(configPath, []byte(configContent), 0644)

	// Load config
	cfg, err := LoadYAMLConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadYAMLConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	// Verify values
	if cfg.API.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", cfg.API.Provider)
	}

	if cfg.API.Key != "test-api-key" {
		t.Errorf("Expected key 'test-api-key', got '%s'", cfg.API.Key)
	}

	if cfg.API.BaseURL != "https://api.test.com/v1" {
		t.Errorf("Expected base_url 'https://api.test.com/v1', got '%s'", cfg.API.BaseURL)
	}

	if cfg.API.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", cfg.API.Model)
	}

	if cfg.API.MaxTokens != 8192 {
		t.Errorf("Expected max_tokens 8192, got %d", cfg.API.MaxTokens)
	}

	if cfg.API.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %f", cfg.API.Temperature)
	}
}

func TestLoadYAMLConfig_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to load config when file doesn't exist
	cfg, err := LoadYAMLConfig(tmpDir)

	// Should not error, just return nil
	if err != nil {
		t.Errorf("Expected no error when config doesn't exist, got: %v", err)
	}

	if cfg != nil {
		t.Error("Expected nil config when file doesn't exist")
	}
}

func TestExpandEnvVars(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple env var",
			input:    "${TEST_VAR}",
			expected: "test-value",
		},
		{
			name:     "Env var in string",
			input:    "prefix-${TEST_VAR}-suffix",
			expected: "prefix-test-value-suffix",
		},
		{
			name:     "No env var",
			input:    "no-vars-here",
			expected: "no-vars-here",
		},
		{
			name:     "Non-existent env var",
			input:    "${NON_EXISTENT_VAR}",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := os.ExpandEnv(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestMergeWithEnv(t *testing.T) {
	yamlCfg := &YAMLConfig{}
	yamlCfg.API.Key = "yaml-key"
	yamlCfg.API.BaseURL = "yaml-url"
	yamlCfg.API.Model = "yaml-model"
	yamlCfg.API.MaxTokens = 4096
	yamlCfg.API.Temperature = 0.5

	envCfg := &Config{
		APIKey:      "env-key",
		BaseURL:     "", // Not set in env
		Model:       "env-model",
		MaxTokens:   8192,
		Temperature: 0.7,
	}

	// Merge - env takes precedence for set values
	merged := MergeWithEnv(yamlCfg, envCfg)

	// Env key should override
	if merged.APIKey != "env-key" {
		t.Errorf("Expected APIKey 'env-key', got '%s'", merged.APIKey)
	}

	// Env BaseURL is empty, so yaml should be used
	if merged.BaseURL != "yaml-url" {
		t.Errorf("Expected BaseURL 'yaml-url', got '%s'", merged.BaseURL)
	}

	// Env model should override
	if merged.Model != "env-model" {
		t.Errorf("Expected Model 'env-model', got '%s'", merged.Model)
	}

	// Env values should override
	if merged.MaxTokens != 8192 {
		t.Errorf("Expected MaxTokens 8192, got %d", merged.MaxTokens)
	}

	if merged.Temperature != 0.7 {
		t.Errorf("Expected Temperature 0.7, got %f", merged.Temperature)
	}
}
