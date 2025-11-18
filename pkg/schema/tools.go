package schema

// ToolDefinition represents a tool definition for the API
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// Common tool schemas
var (
	BashToolSchema = ToolDefinition{
		Name:        "Bash",
		Description: "Executes a bash command with optional timeout",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The command to execute",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Clear, concise description of what this command does",
				},
				"timeout": map[string]interface{}{
					"type":        "number",
					"description": "Optional timeout in milliseconds (max 600000)",
				},
				"run_in_background": map[string]interface{}{
					"type":        "boolean",
					"description": "Set to true to run this command in the background",
				},
			},
			"required": []string{"command"},
		},
	}

	ReadToolSchema = ToolDefinition{
		Name:        "Read",
		Description: "Reads a file from the filesystem",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute path to the file to read",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "The line number to start reading from",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "The number of lines to read",
				},
			},
			"required": []string{"file_path"},
		},
	}

	WriteToolSchema = ToolDefinition{
		Name:        "Write",
		Description: "Writes a file to the filesystem",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute path to the file to write",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The content to write to the file",
				},
			},
			"required": []string{"file_path", "content"},
		},
	}

	EditToolSchema = ToolDefinition{
		Name:        "Edit",
		Description: "Performs exact string replacements in files",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute path to the file to modify",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "The text to replace",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "The text to replace it with",
				},
				"replace_all": map[string]interface{}{
					"type":        "boolean",
					"description": "Replace all occurrences of old_string (default false)",
				},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		},
	}

	GlobToolSchema = ToolDefinition{
		Name:        "Glob",
		Description: "Fast file pattern matching tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The glob pattern to match files against",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The directory to search in",
				},
			},
			"required": []string{"pattern"},
		},
	}

	GrepToolSchema = ToolDefinition{
		Name:        "Grep",
		Description: "A powerful search tool built on ripgrep",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The regular expression pattern to search for",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File or directory to search in",
				},
				"glob": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern to filter files",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "File type to search (js, py, rust, go, java, etc.)",
				},
				"output_mode": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"content", "files_with_matches", "count"},
					"description": "Output mode",
				},
				"-i": map[string]interface{}{
					"type":        "boolean",
					"description": "Case insensitive search",
				},
				"-A": map[string]interface{}{
					"type":        "number",
					"description": "Number of lines to show after each match",
				},
				"-B": map[string]interface{}{
					"type":        "number",
					"description": "Number of lines to show before each match",
				},
				"-C": map[string]interface{}{
					"type":        "number",
					"description": "Number of lines to show before and after each match",
				},
				"multiline": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable multiline mode",
				},
			},
			"required": []string{"pattern"},
		},
	}

	TodoWriteToolSchema = ToolDefinition{
		Name:        "TodoWrite",
		Description: "Create and manage a structured task list",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"todos": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"content": map[string]interface{}{
								"type": "string",
							},
							"status": map[string]interface{}{
								"type": "string",
								"enum": []string{"pending", "in_progress", "completed"},
							},
							"activeForm": map[string]interface{}{
								"type": "string",
							},
						},
						"required": []string{"content", "status", "activeForm"},
					},
				},
			},
			"required": []string{"todos"},
		},
	}
)

// GetAllToolSchemas returns all tool definitions
func GetAllToolSchemas() []ToolDefinition {
	return []ToolDefinition{
		BashToolSchema,
		ReadToolSchema,
		WriteToolSchema,
		EditToolSchema,
		GlobToolSchema,
		GrepToolSchema,
		TodoWriteToolSchema,
		// Add more tools as implemented
	}
}
