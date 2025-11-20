package util

import (
	"fmt"
	"os"
	"strings"
)

// RenderMarkdown renders markdown text for terminal display
// This is a simple implementation - for full features, would use glamour library
func RenderMarkdown(text string) string {
	if !strings.Contains(text, "#") && !strings.Contains(text, "```") && !strings.Contains(text, "*") {
		return text // No markdown detected
	}

	lines := strings.Split(text, "\n")
	var result strings.Builder

	inCodeBlock := false
	codeBlockLang := ""

	for _, line := range lines {
		// Code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				result.WriteString(fmt.Sprintf("\033[0m```\033[0m\n")) // Reset color
				inCodeBlock = false
			} else {
				codeBlockLang = strings.TrimPrefix(line, "```")
				result.WriteString(fmt.Sprintf("\033[90m```%s\033[0m\n", codeBlockLang)) // Gray
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			result.WriteString(fmt.Sprintf("\033[36m%s\033[0m\n", line)) // Cyan for code
			continue
		}

		// Headers
		if strings.HasPrefix(line, "# ") {
			result.WriteString(fmt.Sprintf("\033[1;34m%s\033[0m\n", line)) // Bold blue
		} else if strings.HasPrefix(line, "## ") {
			result.WriteString(fmt.Sprintf("\033[1;36m%s\033[0m\n", line)) // Bold cyan
		} else if strings.HasPrefix(line, "### ") {
			result.WriteString(fmt.Sprintf("\033[1m%s\033[0m\n", line)) // Bold
		} else if strings.HasPrefix(line, "####") {
			result.WriteString(fmt.Sprintf("\033[1m%s\033[0m\n", line)) // Bold
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// Bullet points
			result.WriteString(fmt.Sprintf("\033[33m•\033[0m %s\n", line[2:]))
		} else if len(line) > 0 && line[0] >= '0' && line[0] <= '9' && strings.Contains(line, ". ") {
			// Numbered lists
			parts := strings.SplitN(line, ". ", 2)
			if len(parts) == 2 {
				result.WriteString(fmt.Sprintf("\033[33m%s.\033[0m %s\n", parts[0], parts[1]))
			} else {
				result.WriteString(line + "\n")
			}
		} else {
			// Regular text - handle inline formatting
			formatted := line

			// Bold **text**
			formatted = replaceInlineFormatting(formatted, "**", "\033[1m", "\033[0m")

			// Italic *text*
			formatted = replaceInlineFormatting(formatted, "*", "\033[3m", "\033[0m")

			// Inline code `text`
			formatted = replaceInlineFormatting(formatted, "`", "\033[36m", "\033[0m")

			result.WriteString(formatted + "\n")
		}
	}

	return result.String()
}

// replaceInlineFormatting replaces markdown inline formatting with ANSI codes
func replaceInlineFormatting(text, delimiter, startCode, endCode string) string {
	parts := strings.Split(text, delimiter)
	if len(parts) < 3 {
		return text
	}

	result := parts[0]
	for i := 1; i < len(parts); i += 2 {
		if i+1 < len(parts) {
			result += startCode + parts[i] + endCode + parts[i+1]
		} else {
			result += delimiter + parts[i]
		}
	}

	return result
}

// EnableMarkdownRendering checks if markdown should be rendered
func EnableMarkdownRendering() bool {
	// Check if terminal supports colors
	term := strings.ToLower(os.Getenv("TERM"))
	if term == "dumb" || term == "" {
		return false
	}

	// Check if output is a terminal
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
