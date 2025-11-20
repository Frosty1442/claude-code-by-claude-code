package conversation

import (
	"context"
	"fmt"

	"github.com/claude-code-clone/claude-code-clone/internal/api"
	"github.com/claude-code-clone/claude-code-clone/internal/executor"
	"github.com/claude-code-clone/claude-code-clone/internal/tools"
	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// Manager manages conversation state and flow
type Manager struct {
	client   api.Client
	registry *tools.Registry
	executor *executor.Executor
	history  []schema.Message
	maxTokens int
	systemPrompt string
}

// Config holds conversation manager configuration
type Config struct {
	Client       api.Client
	Registry     *tools.Registry
	SystemPrompt string
	MaxTokens    int
}

// NewManager creates a new conversation manager
func NewManager(cfg Config) *Manager {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 200000
	}

	return &Manager{
		client:       cfg.Client,
		registry:     cfg.Registry,
		executor:     executor.NewExecutor(cfg.Registry),
		history:      []schema.Message{},
		maxTokens:    cfg.MaxTokens,
		systemPrompt: cfg.SystemPrompt,
	}
}

// SendMessage sends a user message and gets a response
func (m *Manager) SendMessage(ctx context.Context, userMessage string) (*Response, error) {
	// Add user message to history
	m.history = append(m.history, schema.NewTextMessage("user", userMessage))

	// Prepare messages with system prompt
	messages := m.prepareMessages()

	// Get tool schemas
	toolSchemas := m.registry.GetSchemas()

	// Execute conversation loop
	return m.conversationLoop(ctx, messages, toolSchemas)
}

// conversationLoop handles the conversation with tool execution
func (m *Manager) conversationLoop(ctx context.Context, messages []schema.Message, toolSchemas []schema.ToolDefinition) (*Response, error) {
	maxIterations := 10 // Prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++

		// Send to API
		resp, err := m.client.SendMessage(ctx, messages, toolSchemas)
		if err != nil {
			return nil, fmt.Errorf("API error: %w", err)
		}

		// Add assistant response to history
		assistantMsg := schema.Message{
			Role:    "assistant",
			Content: resp.Content,
		}
		m.history = append(m.history, assistantMsg)

		// Check if there are tool calls
		toolCalls := assistantMsg.GetToolUses()
		if len(toolCalls) == 0 {
			// No more tool calls, return final response
			return &Response{
				Text:  assistantMsg.GetText(),
				Usage: resp.Usage,
			}, nil
		}

		// Execute tools in parallel (they're independent by default)
		toolResults := m.executor.ExecuteParallel(toolCalls)

		// Convert to content blocks
		resultBlocks := make([]schema.ContentBlock, len(toolResults))
		for i, result := range toolResults {
			resultBlocks[i] = result.ToContentBlock()
		}

		// Add tool results to history
		toolResultMsg := schema.Message{
			Role:    "user",
			Content: resultBlocks,
		}
		m.history = append(m.history, toolResultMsg)

		// Prepare for next iteration
		messages = m.prepareMessages()
	}

	return nil, fmt.Errorf("max iterations reached without final response")
}

// prepareMessages prepares messages for API call
func (m *Manager) prepareMessages() []schema.Message {
	messages := []schema.Message{}

	// Add system prompt if configured
	if m.systemPrompt != "" {
		messages = append(messages, schema.NewTextMessage("system", m.systemPrompt))
	}

	// Add conversation history
	messages = append(messages, m.history...)

	return messages
}

// GetHistory returns the conversation history
func (m *Manager) GetHistory() []schema.Message {
	return m.history
}

// ClearHistory clears the conversation history
func (m *Manager) ClearHistory() {
	m.history = []schema.Message{}
}

// Response represents a conversation response
type Response struct {
	Text  string
	Usage api.Usage
}
