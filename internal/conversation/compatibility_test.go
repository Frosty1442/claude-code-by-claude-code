package conversation

import (
	"testing"
)

func TestParseXMLToolRequests(t *testing.T) {
	text := `Let me read that file for you.

<tool_use name="Read">
<file_path>/home/user/README.md</file_path>
<limit>100</limit>
</tool_use>

I'll execute this now.`

	requests := ParseToolRequestFromText(text)

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Name != "Read" {
		t.Errorf("Expected tool name 'Read', got '%s'", req.Name)
	}

	if req.Parameters["file_path"] != "/home/user/README.md" {
		t.Errorf("Expected file_path '/home/user/README.md', got '%v'", req.Parameters["file_path"])
	}

	if req.Parameters["limit"] != "100" {
		t.Errorf("Expected limit '100', got '%v'", req.Parameters["limit"])
	}
}

func TestParseJSONToolRequests(t *testing.T) {
	text := `I'll use the Glob tool to find those files.

` + "```json\n" + `{
  "tool": "Glob",
  "params": {
    "pattern": "*.go"
  }
}
` + "```\n"

	requests := ParseToolRequestFromText(text)

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Name != "Glob" {
		t.Errorf("Expected tool name 'Glob', got '%s'", req.Name)
	}

	params, ok := req.Parameters["pattern"]
	if !ok || params != "*.go" {
		t.Errorf("Expected pattern '*.go', got '%v'", params)
	}
}

func TestParseReActToolRequests(t *testing.T) {
	text := `Thought: I need to find all Go files in the project.
Action: Glob[pattern="**/*.go"]
Observation: (waiting for tool result)`

	requests := ParseToolRequestFromText(text)

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Name != "Glob" {
		t.Errorf("Expected tool name 'Glob', got '%s'", req.Name)
	}

	if req.Parameters["pattern"] != "**/*.go" {
		t.Errorf("Expected pattern '**/*.go', got '%v'", req.Parameters["pattern"])
	}
}

func TestParseMultipleTools(t *testing.T) {
	text := `I'll do two things:

<tool_use name="Glob">
<pattern>*.go</pattern>
</tool_use>

<tool_use name="Grep">
<pattern>TODO</pattern>
<output_mode>count</output_mode>
</tool_use>

Done!`

	requests := ParseToolRequestFromText(text)

	if len(requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(requests))
	}

	if requests[0].Name != "Glob" {
		t.Errorf("Expected first tool 'Glob', got '%s'", requests[0].Name)
	}

	if requests[1].Name != "Grep" {
		t.Errorf("Expected second tool 'Grep', got '%s'", requests[1].Name)
	}
}

func TestCompatibilitySystemPrompt(t *testing.T) {
	prompt := GetCompatibilitySystemPrompt()

	if prompt == "" {
		t.Error("System prompt should not be empty")
	}

	// Check it contains tool information
	if !contains(prompt, "tool_use") {
		t.Error("System prompt should mention tool_use syntax")
	}

	if !contains(prompt, "Read") || !contains(prompt, "Bash") || !contains(prompt, "Glob") {
		t.Error("System prompt should list available tools")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && len(s) >= len(substr) && (s == substr || findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
