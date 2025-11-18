package schema

import "encoding/json"

// Message represents a conversation message
type Message struct {
	Role    string        `json:"role"`    // "user", "assistant", "system"
	Content []ContentBlock `json:"content"` // Can be string or array of content blocks
}

// ContentBlock represents a piece of message content
type ContentBlock struct {
	Type string `json:"type"` // "text", "tool_use", "tool_result", "image"

	// For text content
	Text string `json:"text,omitempty"`

	// For tool use
	ID         string                 `json:"id,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Input      map[string]interface{} `json:"input,omitempty"`

	// For tool results
	ToolUseID  string `json:"tool_use_id,omitempty"`
	Content    string `json:"content,omitempty"`
	IsError    bool   `json:"is_error,omitempty"`

	// For images
	Source     *ImageSource `json:"source,omitempty"`
}

// ImageSource represents an image source
type ImageSource struct {
	Type      string `json:"type"`       // "base64", "url"
	MediaType string `json:"media_type"` // "image/png", "image/jpeg", etc.
	Data      string `json:"data"`       // base64 data or URL
}

// NewTextMessage creates a text message
func NewTextMessage(role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentBlock{
			{Type: "text", Text: text},
		},
	}
}

// NewToolUseBlock creates a tool use content block
func NewToolUseBlock(id, name string, input map[string]interface{}) ContentBlock {
	return ContentBlock{
		Type:  "tool_use",
		ID:    id,
		Name:  name,
		Input: input,
	}
}

// NewToolResultBlock creates a tool result content block
func NewToolResultBlock(toolUseID, content string, isError bool) ContentBlock {
	return ContentBlock{
		Type:      "tool_result",
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   isError,
	}
}

// GetText extracts text from a message
func (m *Message) GetText() string {
	for _, block := range m.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}

// GetToolUses extracts all tool use blocks
func (m *Message) GetToolUses() []ContentBlock {
	var tools []ContentBlock
	for _, block := range m.Content {
		if block.Type == "tool_use" {
			tools = append(tools, block)
		}
	}
	return tools
}

// MarshalJSON custom marshaling to handle string content shorthand
func (m Message) MarshalJSON() ([]byte, error) {
	type Alias Message

	// If only one text block, use string shorthand
	if len(m.Content) == 1 && m.Content[0].Type == "text" {
		return json.Marshal(&struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			Role:    m.Role,
			Content: m.Content[0].Text,
		})
	}

	return json.Marshal(Alias(m))
}

// UnmarshalJSON custom unmarshaling to handle string content shorthand
func (m *Message) UnmarshalJSON(data []byte) error {
	// Try array of content blocks first
	type Alias Message
	aux := &struct {
		*Alias
		Content json.RawMessage `json:"content"`
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Check if content is a string
	var str string
	if err := json.Unmarshal(aux.Content, &str); err == nil {
		m.Content = []ContentBlock{{Type: "text", Text: str}}
		return nil
	}

	// Otherwise, unmarshal as array of content blocks
	return json.Unmarshal(aux.Content, &m.Content)
}
