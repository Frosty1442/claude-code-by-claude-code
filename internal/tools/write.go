package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// WriteTool writes files
type WriteTool struct {
	workDir string
}

// NewWriteTool creates a new Write tool
func NewWriteTool(workDir string) *WriteTool {
	return &WriteTool{workDir: workDir}
}

// Name returns the tool name
func (t *WriteTool) Name() string {
	return "Write"
}

// Description returns the tool description
func (t *WriteTool) Description() string {
	return "Writes a file to the filesystem"
}

// Schema returns the tool schema
func (t *WriteTool) Schema() schema.ToolDefinition {
	return schema.WriteToolSchema
}

// Validate validates the parameters
func (t *WriteTool) Validate(params map[string]interface{}) error {
	if _, ok := getStringParam(params, "file_path"); !ok {
		return fmt.Errorf("file_path parameter is required")
	}
	if _, ok := getStringParam(params, "content"); !ok {
		return fmt.Errorf("content parameter is required")
	}
	return nil
}

// Execute runs the tool
func (t *WriteTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	filePath, _ := getStringParam(params, "file_path")
	content, _ := getStringParam(params, "content")

	// Resolve absolute path
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(t.workDir, filePath)
	}

	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewErrorResult(fmt.Errorf("failed to create directory: %w", err)), nil
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return NewErrorResult(fmt.Errorf("failed to write file: %w", err)), nil
	}

	return NewSuccessResult(fmt.Sprintf("File created successfully at: %s", filePath)), nil
}
