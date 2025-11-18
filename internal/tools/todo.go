package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// TodoWriteTool manages task lists
type TodoWriteTool struct {
	todos []TodoItem
}

// TodoItem represents a single todo item
type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"activeForm"`
}

// NewTodoWriteTool creates a new TodoWrite tool
func NewTodoWriteTool() *TodoWriteTool {
	return &TodoWriteTool{
		todos: []TodoItem{},
	}
}

// Name returns the tool name
func (t *TodoWriteTool) Name() string {
	return "TodoWrite"
}

// Description returns the tool description
func (t *TodoWriteTool) Description() string {
	return "Create and manage a structured task list"
}

// Schema returns the tool schema
func (t *TodoWriteTool) Schema() schema.ToolDefinition {
	return schema.TodoWriteToolSchema
}

// Validate validates the parameters
func (t *TodoWriteTool) Validate(params map[string]interface{}) error {
	if _, ok := params["todos"]; !ok {
		return fmt.Errorf("todos parameter is required")
	}
	return nil
}

// Execute runs the tool
func (t *TodoWriteTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	todosRaw, ok := params["todos"]
	if !ok {
		return NewErrorResult(fmt.Errorf("todos parameter is required")), nil
	}

	// Parse todos
	todosJSON, err := json.Marshal(todosRaw)
	if err != nil {
		return NewErrorResult(fmt.Errorf("failed to parse todos: %w", err)), nil
	}

	var todos []TodoItem
	if err := json.Unmarshal(todosJSON, &todos); err != nil {
		return NewErrorResult(fmt.Errorf("failed to parse todos: %w", err)), nil
	}

	// Update todos
	t.todos = todos

	// Format output
	output := t.formatTodos()
	return NewSuccessResult(output), nil
}

// formatTodos formats the todo list for display
func (t *TodoWriteTool) formatTodos() string {
	var output strings.Builder
	output.WriteString("Todos have been modified successfully. Ensure that you continue to use the todo list to track your progress. Please proceed with the current tasks if applicable")

	// Could add more detailed output here
	if len(t.todos) > 0 {
		output.WriteString("\n\nCurrent tasks:\n")
		for i, todo := range t.todos {
			status := "○"
			switch todo.Status {
			case "in_progress":
				status = "◐"
			case "completed":
				status = "●"
			}
			output.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, status, todo.Content))
		}
	}

	return output.String()
}

// GetTodos returns the current todo list
func (t *TodoWriteTool) GetTodos() []TodoItem {
	return t.todos
}
