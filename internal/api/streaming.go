package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// StreamingClient extends Client with streaming support
type StreamingClient interface {
	Client
	// StreamMessageWithCallback streams a message and calls callback for each chunk
	StreamMessageWithCallback(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition, callback func(chunk string)) (*Response, error)
}

// OpenAIStreamingClient adds streaming to OpenAIClient
type OpenAIStreamingClient struct {
	*OpenAIClient
}

// NewOpenAIStreamingClient creates a streaming-capable client
func NewOpenAIStreamingClient(cfg Config) (*OpenAIStreamingClient, error) {
	client, err := NewOpenAIClient(cfg)
	if err != nil {
		return nil, err
	}
	return &OpenAIStreamingClient{OpenAIClient: client}, nil
}

// StreamMessageWithCallback streams responses and calls callback for each chunk
func (c *OpenAIStreamingClient) StreamMessageWithCallback(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition, callback func(chunk string)) (*Response, error) {
	// Use the existing StreamMessage
	eventChan, err := c.StreamMessage(ctx, messages, tools)
	if err != nil {
		return nil, err
	}

	var fullContent strings.Builder
	var toolUses []schema.ContentBlock
	var usage Usage

	for event := range eventChan {
		if event.Error != nil {
			return nil, event.Error
		}

		switch event.Type {
		case "content":
			fullContent.WriteString(event.Content)
			if callback != nil {
				callback(event.Content)
			}

		case "tool_use":
			if event.ToolUse != nil {
				toolUses = append(toolUses, *event.ToolUse)
			}

		case "end":
			// Stream complete
		}
	}

	// Build response
	content := []schema.ContentBlock{}
	if fullContent.Len() > 0 {
		content = append(content, schema.ContentBlock{
			Type: "text",
			Text: fullContent.String(),
		})
	}
	content = append(content, toolUses...)

	return &Response{
		ID:         "stream-response",
		Model:      c.config.Model,
		Role:       "assistant",
		Content:    content,
		StopReason: "stop",
		Usage:      usage,
	}, nil
}

// ImprovedSSEReader handles Server-Sent Events properly
type ImprovedSSEReader struct {
	scanner *bufio.Scanner
}

// NewImprovedSSEReader creates a better SSE reader
func NewImprovedSSEReader(r io.Reader) *ImprovedSSEReader {
	return &ImprovedSSEReader{
		scanner: bufio.NewScanner(r),
	}
}

// Next reads the next SSE event
func (r *ImprovedSSEReader) Next() (string, error) {
	var dataLines []string

	for r.scanner.Scan() {
		line := r.scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if len(dataLines) > 0 {
				return strings.Join(dataLines, "\n"), nil
			}
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Parse field
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return "[DONE]", io.EOF
			}
			dataLines = append(dataLines, data)
		}
	}

	if err := r.scanner.Err(); err != nil {
		return "", err
	}

	return "", io.EOF
}

// ParseSSEChunk parses an SSE chunk into a streaming event
func ParseSSEChunk(data string) (*StreamEvent, error) {
	if data == "[DONE]" {
		return &StreamEvent{Type: "end", Done: true}, nil
	}

	var chunk struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Index        int    `json:"index"`
			Delta        json.RawMessage `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse SSE chunk: %w", err)
	}

	if len(chunk.Choices) == 0 {
		return nil, nil
	}

	choice := chunk.Choices[0]

	// Parse delta
	var delta struct {
		Content   *string `json:"content"`
		ToolCalls []struct {
			Index int `json:"index"`
			ID    string `json:"id"`
			Type  string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	}

	if err := json.Unmarshal(choice.Delta, &delta); err != nil {
		return nil, nil
	}

	// Handle content
	if delta.Content != nil && *delta.Content != "" {
		return &StreamEvent{
			Type:    "content",
			Content: *delta.Content,
		}, nil
	}

	// Handle tool calls
	if len(delta.ToolCalls) > 0 {
		for _, tc := range delta.ToolCalls {
			if tc.Function.Name != "" {
				var args map[string]interface{}
				if tc.Function.Arguments != "" {
					json.Unmarshal([]byte(tc.Function.Arguments), &args)
				}
				block := schema.NewToolUseBlock(tc.ID, tc.Function.Name, args)
				return &StreamEvent{
					Type:    "tool_use",
					ToolUse: &block,
				}, nil
			}
		}
	}

	// Handle finish
	if choice.FinishReason != nil && *choice.FinishReason != "" {
		return &StreamEvent{
			Type: "end",
			Done: true,
		}, nil
	}

	return nil, nil
}
