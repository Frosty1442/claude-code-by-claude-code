package conversation

import (
	"context"
	"fmt"
	"strings"

	"github.com/claude-code-clone/claude-code-clone/internal/api"
	"github.com/claude-code-clone/claude-code-clone/internal/tools"
	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// CompatibilityManager handles conversations with models that don't support function calling
type CompatibilityManager struct {
	client   api.Client
	registry *tools.Registry
	history  []schema.Message
	maxTokens int
}

// NewCompatibilityManager creates a manager for non-function-calling models
func NewCompatibilityManager(cfg Config) *CompatibilityManager {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 200000
	}

	return &CompatibilityManager{
		client:    cfg.Client,
		registry:  cfg.Registry,
		history:   []schema.Message{},
		maxTokens: cfg.MaxTokens,
	}
}

// SendMessage sends a message using compatibility mode (text-based tool parsing)
func (m *CompatibilityManager) SendMessage(ctx context.Context, userMessage string) (*Response, error) {
	// Add user message to history
	m.history = append(m.history, schema.NewTextMessage("user", userMessage))

	// Prepare messages with compatibility system prompt
	messages := []schema.Message{
		schema.NewTextMessage("system", GetCompatibilitySystemPrompt()),
	}
	messages = append(messages, m.history...)

	// Execute conversation loop (no native tool schemas)
	return m.compatibilityLoop(ctx, messages)
}

// compatibilityLoop handles the conversation with text-based tool parsing
func (m *CompatibilityManager) compatibilityLoop(ctx context.Context, messages []schema.Message) (*Response, error) {
	maxIterations := 10

	for iteration := 0; iteration < maxIterations; iteration++ {
		// Send to API (no tool schemas)
		resp, err := m.client.SendMessage(ctx, messages, nil)
		if err != nil {
			return nil, fmt.Errorf("API error: %w", err)
		}

		// Get text response
		responseText := ""
		for _, block := range resp.Content {
			if block.Type == "text" {
				responseText += block.Text
			}
		}

		// Add assistant response to history
		m.history = append(m.history, schema.NewTextMessage("assistant", responseText))

		// Parse tool requests from text
		toolRequests := ParseToolRequestFromText(responseText)

		if len(toolRequests) == 0 {
			// No tool requests found, return final response
			// Remove tool request XML from display
			cleanedText := removeToolXML(responseText)
			return &Response{
				Text:  cleanedText,
				Usage: resp.Usage,
			}, nil
		}

		// Execute tools
		toolResults := m.executeCompatibilityTools(toolRequests)

		// Format results as text and add to conversation
		resultsText := formatToolResults(toolResults)
		m.history = append(m.history, schema.NewTextMessage("user", "Tool Results:\n"+resultsText))

		// Prepare for next iteration
		messages = []schema.Message{
			schema.NewTextMessage("system", GetCompatibilitySystemPrompt()),
		}
		messages = append(messages, m.history...)
	}

	return nil, fmt.Errorf("max iterations reached without final response")
}

// executeCompatibilityTools executes tool requests parsed from text
func (m *CompatibilityManager) executeCompatibilityTools(requests []ToolRequest) []ToolExecutionResult {
	results := make([]ToolExecutionResult, len(requests))

	for i, req := range requests {
		toolResult, err := m.registry.Execute(req.Name, req.Parameters)

		results[i] = ToolExecutionResult{
			ToolName: req.Name,
			Output:   toolResult.Output,
			Error:    err,
			IsError:  toolResult.Type == "error",
		}
	}

	return results
}

// ToolExecutionResult represents a tool execution result in compatibility mode
type ToolExecutionResult struct {
	ToolName string
	Output   string
	Error    error
	IsError  bool
}

// formatToolResults formats tool results as readable text
func formatToolResults(results []ToolExecutionResult) string {
	var output strings.Builder

	for _, result := range results {
		output.WriteString(fmt.Sprintf("\n=== %s Tool Result ===\n", result.ToolName))
		if result.IsError || result.Error != nil {
			output.WriteString("Status: ERROR\n")
			if result.Error != nil {
				output.WriteString(fmt.Sprintf("Error: %v\n", result.Error))
			}
		} else {
			output.WriteString("Status: SUCCESS\n")
		}
		output.WriteString(result.Output)
		output.WriteString("\n")
	}

	return output.String()
}

// removeToolXML removes tool request XML tags from text for cleaner display
func removeToolXML(text string) string {
	// Remove <tool_use ...>...</tool_use> blocks
	cleaned := text

	// Simple regex to remove tool_use blocks
	toolPattern := `<tool_use[^>]*>.*?</tool_use>`
	cleaned = strings.TrimSpace(strings.Join(strings.Split(cleaned, toolPattern), ""))

	return cleaned
}

// GetHistory returns the conversation history
func (m *CompatibilityManager) GetHistory() []schema.Message {
	return m.history
}

// ClearHistory clears the conversation history
func (m *CompatibilityManager) ClearHistory() {
	m.history = []schema.Message{}
}
