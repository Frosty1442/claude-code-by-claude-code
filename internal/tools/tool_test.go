package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBashTool(t *testing.T) {
	tool := NewBashTool("/tmp")

	// Test successful command
	result, err := tool.Execute(map[string]interface{}{
		"command": "echo 'Hello World'",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type != "success" && result.Type != "system" {
		t.Errorf("Expected success or system result, got %s", result.Type)
	}
}

func TestReadTool(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!\nLine 2\nLine 3"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewReadTool(tmpDir)

	// Test reading the file
	result, err := tool.Execute(map[string]interface{}{
		"file_path": testFile,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type == "error" {
		t.Errorf("Expected success, got error: %s", result.Output)
	}
}

func TestWriteTool(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := NewWriteTool(tmpDir)

	// Test writing a file
	result, err := tool.Execute(map[string]interface{}{
		"file_path": testFile,
		"content":   "Hello, World!",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type == "error" {
		t.Errorf("Expected success, got error: %s", result.Output)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File was not created")
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != "Hello, World!" {
		t.Errorf("Content mismatch: got %s", string(content))
	}
}

func TestEditTool(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	initialContent := "Hello, World!"

	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewEditTool(tmpDir)

	// Test editing the file
	result, err := tool.Execute(map[string]interface{}{
		"file_path":  testFile,
		"old_string": "World",
		"new_string": "Go",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type == "error" {
		t.Errorf("Expected success, got error: %s", result.Output)
	}

	// Verify content was changed
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "Hello, Go!"
	if string(content) != expected {
		t.Errorf("Content mismatch: expected %s, got %s", expected, string(content))
	}
}

func TestGlobTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	files := []string{"test1.go", "test2.go", "test.txt"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	tool := NewGlobTool(tmpDir)

	// Test glob pattern
	result, err := tool.Execute(map[string]interface{}{
		"pattern": "*.go",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type == "error" {
		t.Errorf("Expected success, got error: %s", result.Output)
	}
}

func TestGrepTool(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!\nThis is a test\nHello again"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewGrepTool(tmpDir)

	// Test grep
	result, err := tool.Execute(map[string]interface{}{
		"pattern":     "Hello",
		"output_mode": "count",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type == "error" {
		t.Errorf("Expected success, got error: %s", result.Output)
	}
}

func TestTodoWriteTool(t *testing.T) {
	tool := NewTodoWriteTool()

	// Test creating todos
	result, err := tool.Execute(map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Test task 1",
				"status":     "pending",
				"activeForm": "Testing task 1",
			},
			map[string]interface{}{
				"content":    "Test task 2",
				"status":     "in_progress",
				"activeForm": "Testing task 2",
			},
		},
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type == "error" {
		t.Errorf("Expected success, got error: %s", result.Output)
	}

	// Verify todos were set
	todos := tool.GetTodos()
	if len(todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(todos))
	}
}

func TestRegistry(t *testing.T) {
	registry := NewDefaultRegistry("/tmp")

	// Test getting all tools
	tools := registry.GetAll()
	if len(tools) == 0 {
		t.Error("Expected tools in registry, got none")
	}

	// Test getting specific tool
	tool, err := registry.Get("Bash")
	if err != nil {
		t.Errorf("Failed to get Bash tool: %v", err)
	}

	if tool.Name() != "Bash" {
		t.Errorf("Expected Bash, got %s", tool.Name())
	}

	// Test getting non-existent tool
	_, err = registry.Get("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}
}
