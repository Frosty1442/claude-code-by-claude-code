package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// EditTool edits files
type EditTool struct {
	workDir string
}

// NewEditTool creates a new Edit tool
func NewEditTool(workDir string) *EditTool {
	return &EditTool{workDir: workDir}
}

// Name returns the tool name
func (t *EditTool) Name() string {
	return "Edit"
}

// Description returns the tool description
func (t *EditTool) Description() string {
	return "Performs exact string replacements in files"
}

// Schema returns the tool schema
func (t *EditTool) Schema() schema.ToolDefinition {
	return schema.EditToolSchema
}

// Validate validates the parameters
func (t *EditTool) Validate(params map[string]interface{}) error {
	if _, ok := getStringParam(params, "file_path"); !ok {
		return fmt.Errorf("file_path parameter is required")
	}
	if _, ok := getStringParam(params, "old_string"); !ok {
		return fmt.Errorf("old_string parameter is required")
	}
	if _, ok := getStringParam(params, "new_string"); !ok {
		return fmt.Errorf("new_string parameter is required")
	}
	return nil
}

// Execute runs the tool
func (t *EditTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	filePath, _ := getStringParam(params, "file_path")
	oldString, _ := getStringParam(params, "old_string")
	newString, _ := getStringParam(params, "new_string")
	replaceAll, _ := getBoolParam(params, "replace_all")

	// Resolve absolute path
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(t.workDir, filePath)
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewErrorResult(fmt.Errorf("File does not exist.")), nil
		}
		return NewErrorResult(err), nil
	}

	fileContent := string(content)

	// Check if old_string exists
	if !strings.Contains(fileContent, oldString) {
		return NewErrorResult(fmt.Errorf("old_string not found in file")), nil
	}

	// Check uniqueness if not replace_all
	if !replaceAll {
		count := strings.Count(fileContent, oldString)
		if count > 1 {
			return NewErrorResult(fmt.Errorf("old_string appears %d times in file. Use replace_all: true or provide a more specific string", count)), nil
		}
	}

	// Perform replacement
	var newContent string
	if replaceAll {
		newContent = strings.ReplaceAll(fileContent, oldString, newString)
	} else {
		newContent = strings.Replace(fileContent, oldString, newString, 1)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return NewErrorResult(fmt.Errorf("failed to write file: %w", err)), nil
	}

	replacements := 1
	if replaceAll {
		replacements = strings.Count(fileContent, oldString)
	}

	return NewSuccessResult(fmt.Sprintf("File edited successfully. Made %d replacement(s).", replacements)), nil
}
