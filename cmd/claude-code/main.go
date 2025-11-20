package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/claude-code-clone/claude-code-clone/internal/api"
	"github.com/claude-code-clone/claude-code-clone/internal/config"
	"github.com/claude-code-clone/claude-code-clone/internal/conversation"
	"github.com/claude-code-clone/claude-code-clone/internal/repl"
	"github.com/claude-code-clone/claude-code-clone/internal/tools"
)

func main() {
	// Check for compatibility mode flag
	compatMode := false
	for _, arg := range os.Args {
		if arg == "--compatibility" || arg == "-compatibility" {
			compatMode = true
		}
		if arg == "--version" || arg == "-version" {
			fmt.Println("Claude Code Clone v1.0.0")
			fmt.Println("Compatibility: OpenAI-compatible APIs")
			return
		}
	}

	if err := run(compatMode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(compatibilityMode bool) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create API client
	apiCfg := api.Config{
		APIKey:      cfg.APIKey,
		BaseURL:     cfg.BaseURL,
		Model:       cfg.Model,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
	}

	client, err := api.NewClient(apiCfg)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Create tool registry
	registry := tools.NewDefaultRegistry(cfg.WorkDir)

	// Create REPL based on mode
	var replInstance *repl.REPL

	if compatibilityMode {
		// Compatibility mode for models without function calling
		compatMgr := conversation.NewCompatibilityManager(conversation.Config{
			Client:   client,
			Registry: registry,
			MaxTokens: 200000,
		})
		replInstance = repl.NewREPLWithCompatibility(compatMgr, cfg.WorkDir, client)
	} else {
		// Standard mode with native function calling
		convMgr := conversation.NewManager(conversation.Config{
			Client:       client,
			Registry:     registry,
			SystemPrompt: cfg.SystemPrompt,
			MaxTokens:    200000,
		})
		replInstance = repl.NewREPL(convMgr, cfg.WorkDir, client)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Shutting down...")
		cancel()
	}()

	// Run REPL
	return replInstance.Run(ctx)
}
