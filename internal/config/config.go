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
	cfg.SystemPrompt = getSystemPrompt()

	return cfg, nil
}

// getSystemPrompt returns the default system prompt
func getSystemPrompt() string {
	return `You are an AI coding assistant with access to various tools for file operations, command execution, and code analysis.

You have access to the following tools:
- Bash: Execute shell commands
- Read: Read files from the filesystem
- Write: Create or overwrite files
- Edit: Make precise edits to existing files
- Glob: Find files by pattern
- Grep: Search file contents with regex
- TodoWrite: Manage task lists
- Git: Git operations (status, diff, commit, push, pull, branch)
- WebFetch: Fetch content from URLs

Guidelines:
1. Always use absolute paths for file operations
2. Use the appropriate tool for each task
3. Provide clear, concise responses
4. Execute multiple independent tools in parallel when possible
5. Validate file paths before operations
6. Handle errors gracefully

When asked to perform tasks:
- Break down complex tasks into steps
- Use TodoWrite to track progress on multi-step tasks
- Execute tools systematically
- Verify results before proceeding

Your goal is to help users with software development tasks efficiently and accurately.`
}
