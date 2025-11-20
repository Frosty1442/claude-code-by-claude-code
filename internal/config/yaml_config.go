package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// YAMLConfig represents configuration from .claude/config.yaml
type YAMLConfig struct {
	API struct {
		Provider    string  `yaml:"provider"`
		Key         string  `yaml:"key"`
		BaseURL     string  `yaml:"base_url"`
		Model       string  `yaml:"model"`
		MaxTokens   int     `yaml:"max_tokens"`
		Temperature float64 `yaml:"temperature"`
	} `yaml:"api"`

	Tools struct {
		Bash struct {
			Timeout int  `yaml:"timeout"`
			Sandbox bool `yaml:"sandbox"`
		} `yaml:"bash"`
	} `yaml:"tools"`

	Features struct {
		Streaming    bool `yaml:"streaming"`
		ParallelExec bool `yaml:"parallel_exec"`
	} `yaml:"features"`

	UI struct {
		Markdown bool   `yaml:"markdown"`
		Theme    string `yaml:"theme"`
	} `yaml:"ui"`
}

// LoadYAMLConfig loads configuration from .claude/config.yaml if it exists
func LoadYAMLConfig(workDir string) (*YAMLConfig, error) {
	configPath := filepath.Join(workDir, ".claude", "config.yaml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // No config file, return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg YAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Expand environment variables in key
	if cfg.API.Key != "" {
		cfg.API.Key = os.ExpandEnv(cfg.API.Key)
	}

	return &cfg, nil
}

// MergeWithEnv merges YAML config with environment variables
// Environment variables take precedence
func MergeWithEnv(yamlCfg *YAMLConfig, envCfg *Config) *Config {
	result := envCfg

	if yamlCfg == nil {
		return result
	}

	// Use YAML values if env not set
	if result.BaseURL == "" && yamlCfg.API.BaseURL != "" {
		result.BaseURL = yamlCfg.API.BaseURL
	}

	if result.Model == "" && yamlCfg.API.Model != "" {
		result.Model = yamlCfg.API.Model
	}

	if result.MaxTokens == 0 && yamlCfg.API.MaxTokens > 0 {
		result.MaxTokens = yamlCfg.API.MaxTokens
	}

	if result.Temperature == 0 && yamlCfg.API.Temperature > 0 {
		result.Temperature = yamlCfg.API.Temperature
	}

	// API key from YAML if env not set
	if result.APIKey == "" && yamlCfg.API.Key != "" {
		result.APIKey = yamlCfg.API.Key
	}

	return result
}

// CreateDefaultConfigFile creates a default .claude/config.yaml
func CreateDefaultConfigFile(workDir string) error {
	claudeDir := filepath.Join(workDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(claudeDir, "config.yaml")

	// Don't overwrite existing config
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists")
	}

	defaultConfig := `# Claude Code Clone Configuration

api:
  provider: openai  # openai, anthropic, custom
  key: ${OPENAI_API_KEY}  # or set directly (not recommended)
  base_url: https://api.openai.com/v1
  model: gpt-4
  max_tokens: 4096
  temperature: 0.7

tools:
  bash:
    timeout: 300000  # milliseconds
    sandbox: false   # restrict to working directory

features:
  streaming: true
  parallel_exec: true  # execute independent tools in parallel

ui:
  markdown: true  # render markdown in output
  theme: auto     # auto, light, dark
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
