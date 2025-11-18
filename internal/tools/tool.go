package tools

import (
	"fmt"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// Tool represents a tool that can be executed
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Schema returns the JSON schema for tool parameters
	Schema() schema.ToolDefinition

	// Execute runs the tool with the given parameters
	Execute(params map[string]interface{}) (*ToolResult, error)

	// Validate checks if parameters are valid
	Validate(params map[string]interface{}) error
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Output  string
	Error   error
	Type    string // "success", "error", "system"
}

// NewSuccessResult creates a success result
func NewSuccessResult(output string) *ToolResult {
	return &ToolResult{
		Output: output,
		Type:   "success",
	}
}

// NewErrorResult creates an error result
func NewErrorResult(err error) *ToolResult {
	return &ToolResult{
		Output: err.Error(),
		Error:  err,
		Type:   "error",
	}
}

// NewSystemResult creates a system result
func NewSystemResult(output string) *ToolResult {
	return &ToolResult{
		Output: output,
		Type:   "system",
	}
}

// Registry holds all available tools
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

// GetAll returns all registered tools
func (r *Registry) GetAll() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetSchemas returns schemas for all tools
func (r *Registry) GetSchemas() []schema.ToolDefinition {
	schemas := make([]schema.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		schemas = append(schemas, tool.Schema())
	}
	return schemas
}

// Execute executes a tool by name
func (r *Registry) Execute(name string, params map[string]interface{}) (*ToolResult, error) {
	tool, err := r.Get(name)
	if err != nil {
		return NewErrorResult(err), err
	}

	if err := tool.Validate(params); err != nil {
		return NewErrorResult(fmt.Errorf("validation failed: %w", err)), err
	}

	return tool.Execute(params)
}

// NewDefaultRegistry creates a registry with all default tools
func NewDefaultRegistry(workDir string) *Registry {
	r := NewRegistry()

	// Register all tools
	r.Register(NewBashTool(workDir))
	r.Register(NewReadTool(workDir))
	r.Register(NewWriteTool(workDir))
	r.Register(NewEditTool(workDir))
	r.Register(NewGlobTool(workDir))
	r.Register(NewGrepTool(workDir))
	r.Register(NewTodoWriteTool())

	return r
}

// getStringParam extracts a string parameter
func getStringParam(params map[string]interface{}, key string) (string, bool) {
	val, ok := params[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// getIntParam extracts an int parameter
func getIntParam(params map[string]interface{}, key string) (int, bool) {
	val, ok := params[key]
	if !ok {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// getBoolParam extracts a bool parameter
func getBoolParam(params map[string]interface{}, key string) (bool, bool) {
	val, ok := params[key]
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}
