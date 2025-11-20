package repl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/claude-code-clone/claude-code-clone/internal/conversation"
)

// REPL represents the interactive REPL
type REPL struct {
	manager       *conversation.Manager
	compatManager *conversation.CompatibilityManager
	reader        *bufio.Reader
	workDir       string
	client        interface{} // Stores the API client for model switching
}

// NewREPL creates a new REPL with standard manager
func NewREPL(manager *conversation.Manager, workDir string, client interface{}) *REPL {
	return &REPL{
		manager: manager,
		reader:  bufio.NewReader(os.Stdin),
		workDir: workDir,
		client:  client,
	}
}

// NewREPLWithCompatibility creates a new REPL with compatibility manager
func NewREPLWithCompatibility(manager *conversation.CompatibilityManager, workDir string, client interface{}) *REPL {
	return &REPL{
		compatManager: manager,
		reader:        bufio.NewReader(os.Stdin),
		workDir:       workDir,
		client:        client,
	}
}

// Run starts the REPL
func (r *REPL) Run(ctx context.Context) error {
	fmt.Println("Claude Code Clone - Interactive AI Assistant")

	if r.compatManager != nil {
		fmt.Println("Mode: Compatibility (for models without native function calling)")
		fmt.Println("The model will use XML-based tool requests")
	}

	fmt.Println("Type your message and press Enter. Type 'exit' to quit, 'clear' to clear history.")
	fmt.Println(strings.Repeat("-", 60))

	for {
		// Print prompt
		fmt.Print("\n> ")

		// Read input
		input, err := r.reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		// Handle commands
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return nil
		}

		if input == "clear" {
			if r.compatManager != nil {
				r.compatManager.ClearHistory()
			} else if r.manager != nil {
				r.manager.ClearHistory()
			}
			fmt.Println("History cleared.")
			continue
		}

		if input == "help" {
			r.printHelp()
			continue
		}

		// Handle model commands
		if strings.HasPrefix(input, "/model ") {
			modelName := strings.TrimSpace(strings.TrimPrefix(input, "/model"))
			r.switchModel(modelName)
			continue
		}

		if input == "/model" || input == "/models" {
			r.showCurrentModel()
			continue
		}

		// Send message
		fmt.Println()

		var resp *conversation.Response
		if r.compatManager != nil {
			resp, err = r.compatManager.SendMessage(ctx, input)
		} else {
			resp, err = r.manager.SendMessage(ctx, input)
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Display response
		fmt.Println(resp.Text)
		fmt.Printf("\n[Tokens: Input=%d, Output=%d]\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}
}

// printHelp displays help information
func (r *REPL) printHelp() {
	fmt.Println(`Available commands:
  help          - Show this help message
  clear         - Clear conversation history
  /model        - Show current model
  /model <name> - Switch to a different model
  exit          - Exit the program

You can ask the assistant to:
  - Read, write, and edit files
  - Execute bash commands
  - Search files by pattern (glob) or content (grep)
  - Manage todo lists
  - Git operations (status, diff, commit, push)
  - Fetch web content
  - And much more!

Example queries:
  "Read the README.md file"
  "Search for all .go files"
  "Find TODO comments in the code"
  "Create a new file called test.txt with some content"
  "Show git status and recent commits"

Common models:
  gpt-4, gpt-4-turbo-preview, gpt-3.5-turbo
  claude-3-opus-20240229, claude-3-sonnet-20240229
  mistral-medium, mistral-large-latest
  Or any locally hosted model endpoint`)

	if r.compatManager != nil {
		fmt.Println(`Compatibility Mode Active:
This mode works with models that don't support native function calling.
The model will request tools using XML syntax like:

  <tool_use name="Read">
  <file_path>/path/to/file</file_path>
  </tool_use>

This enables smaller models (TinyLlama, Mistral 7B, etc.) to use tools!`)
	}
}

// switchModel switches to a different model
func (r *REPL) switchModel(modelName string) {
	if r.client == nil {
		fmt.Println("Error: Client not available for model switching")
		return
	}

	// Try to cast to interface with SetModel method
	type ModelSetter interface {
		SetModel(string)
		GetModel() string
	}

	if setter, ok := r.client.(ModelSetter); ok {
		oldModel := setter.GetModel()
		setter.SetModel(modelName)
		fmt.Printf("Switched model from '%s' to '%s'\n", oldModel, modelName)
		fmt.Println("Note: Conversation history is preserved")
	} else {
		fmt.Println("Error: Current client doesn't support model switching")
	}
}

// showCurrentModel displays the current model
func (r *REPL) showCurrentModel() {
	if r.client == nil {
		fmt.Println("Error: Client not available")
		return
	}

	type ModelGetter interface {
		GetModel() string
	}

	if getter, ok := r.client.(ModelGetter); ok {
		fmt.Printf("Current model: %s\n", getter.GetModel())
		fmt.Println("\nTo switch models, use: /model <model-name>")
		fmt.Println("Example: /model gpt-4-turbo-preview")
	} else {
		fmt.Println("Error: Current client doesn't support model queries")
	}
}
