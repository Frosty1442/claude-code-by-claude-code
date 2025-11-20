package util

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_Headers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "H1 header",
			input:    "# Heading 1",
			contains: "Heading 1",
		},
		{
			name:     "H2 header",
			input:    "## Heading 2",
			contains: "Heading 2",
		},
		{
			name:     "H3 header",
			input:    "### Heading 3",
			contains: "Heading 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderMarkdown(tt.input)
			if !strings.Contains(output, tt.contains) {
				t.Errorf("Expected output to contain '%s'", tt.contains)
			}
			// Should contain ANSI codes
			if !strings.Contains(output, "\033[") {
				t.Error("Expected ANSI escape codes in output")
			}
		})
	}
}

func TestRenderMarkdown_CodeBlocks(t *testing.T) {
	input := "```go\nfunc main() {\n}\n```"
	output := RenderMarkdown(input)

	// Should contain the code
	if !strings.Contains(output, "func main()") {
		t.Error("Expected code content in output")
	}

	// Should contain ANSI codes for styling
	if !strings.Contains(output, "\033[") {
		t.Error("Expected ANSI escape codes in output")
	}
}

func TestRenderMarkdown_InlineCode(t *testing.T) {
	input := "This is `inline code` in text"
	output := RenderMarkdown(input)

	// Should contain the text
	if !strings.Contains(output, "inline code") {
		t.Error("Expected inline code in output")
	}

	if !strings.Contains(output, "This is") {
		t.Error("Expected regular text in output")
	}
}

func TestRenderMarkdown_Lists(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Unordered list",
			input: "- Item 1\n- Item 2\n- Item 3",
		},
		{
			name:  "Ordered list",
			input: "1. First\n2. Second\n3. Third",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderMarkdown(tt.input)
			if output == "" {
				t.Error("Expected non-empty output")
			}
		})
	}
}

func TestRenderMarkdown_Bold(t *testing.T) {
	input := "This is **bold text** in a sentence"
	output := RenderMarkdown(input)

	// Should contain the text
	if !strings.Contains(output, "bold text") {
		t.Error("Expected bold text in output")
	}

	// Should contain bold ANSI codes
	if !strings.Contains(output, "\033[1m") {
		t.Error("Expected bold ANSI codes in output")
	}
}

func TestRenderMarkdown_Italic(t *testing.T) {
	input := "This is *italic text* in a sentence"
	output := RenderMarkdown(input)

	// Should contain the text
	if !strings.Contains(output, "italic text") {
		t.Error("Expected italic text in output")
	}

	// Should contain italic ANSI codes
	if !strings.Contains(output, "\033[3m") {
		t.Error("Expected italic ANSI codes in output")
	}
}

func TestRenderMarkdown_Links(t *testing.T) {
	input := "Check out [this link](https://example.com)"
	output := RenderMarkdown(input)

	// Should contain link text
	if !strings.Contains(output, "this link") {
		t.Error("Expected link text in output")
	}

	// Should contain URL
	if !strings.Contains(output, "https://example.com") {
		t.Error("Expected URL in output")
	}
}

func TestRenderMarkdown_PlainText(t *testing.T) {
	input := "This is plain text with no markdown"
	output := RenderMarkdown(input)

	// Should contain the text
	if !strings.Contains(output, "plain text") {
		t.Error("Expected plain text in output")
	}
}

func TestRenderMarkdown_Mixed(t *testing.T) {
	input := `# Title

This is a paragraph with **bold** and *italic* text.

## Section

- List item 1
- List item 2

` + "```go\ncode here\n```"

	output := RenderMarkdown(input)

	// Should contain all elements
	if !strings.Contains(output, "Title") {
		t.Error("Expected title in output")
	}

	if !strings.Contains(output, "bold") {
		t.Error("Expected bold text in output")
	}

	if !strings.Contains(output, "italic") {
		t.Error("Expected italic text in output")
	}

	if !strings.Contains(output, "Section") {
		t.Error("Expected section header in output")
	}

	if !strings.Contains(output, "List item") {
		t.Error("Expected list items in output")
	}

	if !strings.Contains(output, "code here") {
		t.Error("Expected code block in output")
	}
}

func TestRenderMarkdown_EmptyString(t *testing.T) {
	output := RenderMarkdown("")
	if output != "" {
		t.Error("Expected empty output for empty input")
	}
}
