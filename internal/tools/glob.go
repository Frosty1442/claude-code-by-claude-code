package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// GlobTool finds files by pattern
type GlobTool struct {
	workDir string
}

// NewGlobTool creates a new Glob tool
func NewGlobTool(workDir string) *GlobTool {
	return &GlobTool{workDir: workDir}
}

// Name returns the tool name
func (t *GlobTool) Name() string {
	return "Glob"
}

// Description returns the tool description
func (t *GlobTool) Description() string {
	return "Fast file pattern matching tool"
}

// Schema returns the tool schema
func (t *GlobTool) Schema() schema.ToolDefinition {
	return schema.GlobToolSchema
}

// Validate validates the parameters
func (t *GlobTool) Validate(params map[string]interface{}) error {
	if _, ok := getStringParam(params, "pattern"); !ok {
		return fmt.Errorf("pattern parameter is required")
	}
	return nil
}

// Execute runs the tool
func (t *GlobTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	pattern, _ := getStringParam(params, "pattern")
	searchPath, ok := getStringParam(params, "path")
	if !ok {
		searchPath = t.workDir
	}

	// Resolve absolute path
	if !filepath.IsAbs(searchPath) {
		searchPath = filepath.Join(t.workDir, searchPath)
	}

	// Handle ** patterns with recursive walk
	var matches []string

	if strings.Contains(pattern, "**") {
		matches = t.recursiveGlob(searchPath, pattern)
	} else {
		// Simple glob
		fullPattern := filepath.Join(searchPath, pattern)
		var err error
		matches, err = filepath.Glob(fullPattern)
		if err != nil {
			return NewErrorResult(err), nil
		}
	}

	if len(matches) == 0 {
		return NewSuccessResult("No files matched the pattern."), nil
	}

	// Sort by modification time (newest first)
	sort.Slice(matches, func(i, j int) bool {
		infoI, errI := os.Stat(matches[i])
		infoJ, errJ := os.Stat(matches[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d file(s):\n", len(matches)))
	for _, match := range matches {
		relPath, _ := filepath.Rel(t.workDir, match)
		info, err := os.Stat(match)
		if err == nil {
			modTime := info.ModTime().Format(time.RFC3339)
			output.WriteString(fmt.Sprintf("  %s (modified: %s)\n", relPath, modTime))
		} else {
			output.WriteString(fmt.Sprintf("  %s\n", relPath))
		}
	}

	return NewSuccessResult(output.String()), nil
}

// recursiveGlob handles ** patterns
func (t *GlobTool) recursiveGlob(root, pattern string) []string {
	var matches []string

	// Walk the directory tree
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden files and .git directory
		if strings.HasPrefix(filepath.Base(path), ".") && filepath.Base(path) != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Simple pattern matching
		if matchesPattern(path, root, pattern) {
			matches = append(matches, path)
		}

		return nil
	})

	return matches
}

// matchesPattern checks if a path matches a glob pattern with **
func matchesPattern(path, root, pattern string) bool {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	// Handle ** pattern
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")

		// Simple heuristic: check if the end matches
		if len(parts) >= 2 {
			suffix := strings.TrimPrefix(parts[1], "/")
			matched, _ := filepath.Match(suffix, filepath.Base(relPath))
			return matched
		}
	}

	// Simple glob match
	matched, _ := filepath.Match(pattern, relPath)
	return matched
}
