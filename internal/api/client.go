package api

import (
	"context"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// Client is the interface for LLM API clients
type Client interface {
	// SendMessage sends a message and returns the response
	SendMessage(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (*Response, error)

	// StreamMessage sends a message and streams the response
	StreamMessage(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (<-chan StreamEvent, error)

	// GetModel returns the current model being used
	GetModel() string

	// SetModel sets the model to use
	SetModel(model string)
}

// Response represents an API response
type Response struct {
	ID           string                `json:"id"`
	Model        string                `json:"model"`
	Role         string                `json:"role"`
	Content      []schema.ContentBlock `json:"content"`
	StopReason   string                `json:"stop_reason"`
	Usage        Usage                 `json:"usage"`
}

// Usage represents token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type    string // "start", "content", "tool_use", "end", "error"
	Content string
	ToolUse *schema.ContentBlock
	Error   error
	Done    bool
}

// Config holds API client configuration
type Config struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
	Timeout     int // seconds
}

// NewClient creates a new API client based on configuration
func NewClient(cfg Config) (Client, error) {
	// Default to OpenAI-compatible client
	return NewOpenAIClient(cfg)
}
