package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// OpenAIClient implements the Client interface for OpenAI-compatible endpoints
type OpenAIClient struct {
	config     Config
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI-compatible client
func NewOpenAIClient(cfg Config) (*OpenAIClient, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 300 // 5 minutes default
	}

	return &OpenAIClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}, nil
}

// GetModel returns the current model
func (c *OpenAIClient) GetModel() string {
	return c.config.Model
}

// SetModel sets the model to use
func (c *OpenAIClient) SetModel(model string) {
	c.config.Model = model
}

// openAIRequest represents the OpenAI API request format
type openAIRequest struct {
	Model       string                  `json:"model"`
	Messages    []schema.Message        `json:"messages"`
	Tools       []openAIToolDefinition  `json:"tools,omitempty"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Temperature float64                 `json:"temperature,omitempty"`
	Stream      bool                    `json:"stream,omitempty"`
}

// openAIToolDefinition represents OpenAI tool format
type openAIToolDefinition struct {
	Type     string                 `json:"type"`
	Function openAIFunctionDefinition `json:"function"`
}

// openAIFunctionDefinition represents OpenAI function format
type openAIFunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// openAIResponse represents the OpenAI API response format
type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int              `json:"index"`
		Message      openAIMessage    `json:"message"`
		FinishReason string           `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// openAIMessage represents an OpenAI message
type openAIMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
}

// openAIToolCall represents an OpenAI tool call
type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// SendMessage sends a message and returns the response
func (c *OpenAIClient) SendMessage(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (*Response, error) {
	// Convert tools to OpenAI format
	openAITools := make([]openAIToolDefinition, len(tools))
	for i, tool := range tools {
		openAITools[i] = openAIToolDefinition{
			Type: "function",
			Function: openAIFunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		}
	}

	// Convert messages to OpenAI format
	openAIMessages := convertToOpenAIMessages(messages)

	reqBody := openAIRequest{
		Model:       c.config.Model,
		Messages:    openAIMessages,
		Tools:       openAITools,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert OpenAI response to our format
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openAIResp.Choices[0]
	content := convertFromOpenAIMessage(choice.Message)

	return &Response{
		ID:         openAIResp.ID,
		Model:      openAIResp.Model,
		Role:       "assistant",
		Content:    content,
		StopReason: choice.FinishReason,
		Usage: Usage{
			InputTokens:  openAIResp.Usage.PromptTokens,
			OutputTokens: openAIResp.Usage.CompletionTokens,
		},
	}, nil
}

// StreamMessage sends a message and streams the response
func (c *OpenAIClient) StreamMessage(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (<-chan StreamEvent, error) {
	eventChan := make(chan StreamEvent, 10)

	// Convert tools to OpenAI format
	openAITools := make([]openAIToolDefinition, len(tools))
	for i, tool := range tools {
		openAITools[i] = openAIToolDefinition{
			Type: "function",
			Function: openAIFunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		}
	}

	// Convert messages to OpenAI format
	openAIMessages := convertToOpenAIMessages(messages)

	reqBody := openAIRequest{
		Model:       c.config.Model,
		Messages:    openAIMessages,
		Tools:       openAITools,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		close(eventChan)
		return eventChan, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		close(eventChan)
		return eventChan, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	go func() {
		defer close(eventChan)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			eventChan <- StreamEvent{Type: "error", Error: fmt.Errorf("failed to send request: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			eventChan <- StreamEvent{Type: "error", Error: fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))}
			return
		}

		// Parse SSE stream
		eventChan <- StreamEvent{Type: "start"}

		reader := NewSSEReader(resp.Body)
		for {
			event, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				eventChan <- StreamEvent{Type: "error", Error: err}
				return
			}

			if event == "[DONE]" {
				break
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content   string `json:"content,omitempty"`
						ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(event), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					eventChan <- StreamEvent{Type: "content", Content: delta.Content}
				}
				// Handle tool calls if present
				for _, toolCall := range delta.ToolCalls {
					var args map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err == nil {
						block := schema.NewToolUseBlock(toolCall.ID, toolCall.Function.Name, args)
						eventChan <- StreamEvent{Type: "tool_use", ToolUse: &block}
					}
				}
			}
		}

		eventChan <- StreamEvent{Type: "end", Done: true}
	}()

	return eventChan, nil
}

// convertToOpenAIMessages converts our message format to OpenAI format
func convertToOpenAIMessages(messages []schema.Message) []schema.Message {
	// For now, return as-is since formats are similar
	// In production, would need more sophisticated conversion
	return messages
}

// convertFromOpenAIMessage converts OpenAI message to our content block format
func convertFromOpenAIMessage(msg openAIMessage) []schema.ContentBlock {
	var blocks []schema.ContentBlock

	if msg.Content != "" {
		blocks = append(blocks, schema.ContentBlock{
			Type: "text",
			Text: msg.Content,
		})
	}

	for _, toolCall := range msg.ToolCalls {
		var input map[string]interface{}
		json.Unmarshal([]byte(toolCall.Function.Arguments), &input)

		blocks = append(blocks, schema.ContentBlock{
			Type:  "tool_use",
			ID:    toolCall.ID,
			Name:  toolCall.Function.Name,
			Input: input,
		})
	}

	return blocks
}

// SSEReader reads Server-Sent Events
type SSEReader struct {
	reader io.Reader
}

// NewSSEReader creates a new SSE reader
func NewSSEReader(r io.Reader) *SSEReader {
	return &SSEReader{reader: r}
}

// Next reads the next SSE event
func (r *SSEReader) Next() (string, error) {
	buf := make([]byte, 8192)
	n, err := r.reader.Read(buf)
	if err != nil {
		return "", err
	}

	data := string(buf[:n])
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			return strings.TrimPrefix(line, "data: "), nil
		}
	}

	return "", nil
}
