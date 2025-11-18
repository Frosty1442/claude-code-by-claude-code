package tools

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// ReadTool reads files
type ReadTool struct {
	workDir string
}

// NewReadTool creates a new Read tool
func NewReadTool(workDir string) *ReadTool {
	return &ReadTool{workDir: workDir}
}

// Name returns the tool name
func (t *ReadTool) Name() string {
	return "Read"
}

// Description returns the tool description
func (t *ReadTool) Description() string {
	return "Reads a file from the filesystem"
}

// Schema returns the tool schema
func (t *ReadTool) Schema() schema.ToolDefinition {
	return schema.ReadToolSchema
}

// Validate validates the parameters
func (t *ReadTool) Validate(params map[string]interface{}) error {
	if _, ok := getStringParam(params, "file_path"); !ok {
		return fmt.Errorf("file_path parameter is required")
	}
	return nil
}

// Execute runs the tool
func (t *ReadTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	filePath, _ := getStringParam(params, "file_path")
	offset, hasOffset := getIntParam(params, "offset")
	limit, hasLimit := getIntParam(params, "limit")

	// Resolve absolute path
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(t.workDir, filePath)
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewErrorResult(fmt.Errorf("File does not exist.")), nil
		}
		return NewErrorResult(err), nil
	}

	// Check if it's a directory
	if info.IsDir() {
		return NewErrorResult(fmt.Errorf("Path is a directory, not a file. Use Bash tool with ls command.")), nil
	}

	// Handle images (return base64)
	ext := strings.ToLower(filepath.Ext(filePath))
	if isImageFile(ext) {
		return t.readImage(filePath)
	}

	// Read text file
	return t.readTextFile(filePath, offset, limit, hasOffset, hasLimit)
}

// readTextFile reads a text file with optional offset and limit
func (t *ReadTool) readTextFile(filePath string, offset, limit int, hasOffset, hasLimit bool) (*ToolResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return NewErrorResult(err), nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	lineNum := 1

	// Set default limit
	if !hasLimit {
		limit = 2000
	}
	if !hasOffset {
		offset = 1
	}

	for scanner.Scan() {
		if lineNum >= offset {
			line := scanner.Text()
			// Truncate long lines
			if len(line) > 2000 {
				line = line[:2000] + "... [truncated]"
			}
			lines = append(lines, fmt.Sprintf("%5d\t%s", lineNum, line))

			if len(lines) >= limit {
				break
			}
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return NewErrorResult(err), nil
	}

	if len(lines) == 0 {
		return NewSystemResult("File is empty or has no content in the specified range."), nil
	}

	return NewSuccessResult(strings.Join(lines, "\n")), nil
}

// readImage reads an image file and returns base64
func (t *ReadTool) readImage(filePath string) (*ToolResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return NewErrorResult(err), nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	ext := strings.ToLower(filepath.Ext(filePath))
	mediaType := "image/" + strings.TrimPrefix(ext, ".")

	output := fmt.Sprintf("[Image file: %s, type: %s, size: %d bytes]\nBase64: %s",
		filepath.Base(filePath), mediaType, len(data), encoded[:min(100, len(encoded))]+"...")

	return NewSuccessResult(output), nil
}

// isImageFile checks if the file extension is an image
func isImageFile(ext string) bool {
	imageExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".bmp": true, ".webp": true, ".svg": true,
	}
	return imageExts[ext]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
