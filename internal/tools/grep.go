package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// GrepTool searches file contents
type GrepTool struct {
	workDir string
}

// NewGrepTool creates a new Grep tool
func NewGrepTool(workDir string) *GrepTool {
	return &GrepTool{workDir: workDir}
}

// Name returns the tool name
func (t *GrepTool) Name() string {
	return "Grep"
}

// Description returns the tool description
func (t *GrepTool) Description() string {
	return "A powerful search tool built on ripgrep"
}

// Schema returns the tool schema
func (t *GrepTool) Schema() schema.ToolDefinition {
	return schema.GrepToolSchema
}

// Validate validates the parameters
func (t *GrepTool) Validate(params map[string]interface{}) error {
	if _, ok := getStringParam(params, "pattern"); !ok {
		return fmt.Errorf("pattern parameter is required")
	}
	return nil
}

// Execute runs the tool
func (t *GrepTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	pattern, _ := getStringParam(params, "pattern")
	searchPath, ok := getStringParam(params, "path")
	if !ok {
		searchPath = t.workDir
	}

	outputMode, ok := getStringParam(params, "output_mode")
	if !ok {
		outputMode = "files_with_matches"
	}

	caseInsensitive, _ := getBoolParam(params, "-i")
	contextAfter, _ := getIntParam(params, "-A")
	contextBefore, _ := getIntParam(params, "-B")
	contextBoth, hasC := getIntParam(params, "-C")
	if hasC {
		contextAfter = contextBoth
		contextBefore = contextBoth
	}

	// Resolve absolute path
	if !filepath.IsAbs(searchPath) {
		searchPath = filepath.Join(t.workDir, searchPath)
	}

	// Compile regex
	flags := ""
	if caseInsensitive {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		return NewErrorResult(fmt.Errorf("invalid regex pattern: %w", err)), nil
	}

	// Determine glob pattern
	globPattern, _ := getStringParam(params, "glob")
	fileType, _ := getStringParam(params, "type")

	// Search files
	results := t.searchFiles(searchPath, re, globPattern, fileType, outputMode, contextBefore, contextAfter)

	if len(results) == 0 {
		return NewSuccessResult("No matches found."), nil
	}

	// Format output based on mode
	output := t.formatOutput(results, outputMode)
	return NewSuccessResult(output), nil
}

// searchResult represents a search result
type searchResult struct {
	File    string
	Line    int
	Content string
	Before  []string
	After   []string
}

// searchFiles searches for pattern in files
func (t *GrepTool) searchFiles(searchPath string, re *regexp.Regexp, globPattern, fileType, outputMode string, contextBefore, contextAfter int) []searchResult {
	var results []searchResult
	filesMatched := make(map[string]bool)

	// Walk directory
	filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(filepath.Base(path), ".") && path != searchPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Apply glob filter
		if globPattern != "" {
			matched, _ := filepath.Match(globPattern, filepath.Base(path))
			if !matched {
				return nil
			}
		}

		// Apply type filter
		if fileType != "" {
			ext := filepath.Ext(path)
			if !matchesFileType(ext, fileType) {
				return nil
			}
		}

		// Search file
		fileResults := t.searchFile(path, re, contextBefore, contextAfter)
		if len(fileResults) > 0 {
			filesMatched[path] = true
			results = append(results, fileResults...)
		}

		return nil
	})

	return results
}

// searchFile searches a single file
func (t *GrepTool) searchFile(filePath string, re *regexp.Regexp, contextBefore, contextAfter int) []searchResult {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var results []searchResult
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var recentLines []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Keep recent lines for context
		if contextBefore > 0 {
			recentLines = append(recentLines, line)
			if len(recentLines) > contextBefore {
				recentLines = recentLines[1:]
			}
		}

		// Check for match
		if re.MatchString(line) {
			result := searchResult{
				File:    filePath,
				Line:    lineNum,
				Content: line,
			}

			// Add before context
			if contextBefore > 0 && len(recentLines) > 1 {
				result.Before = make([]string, len(recentLines)-1)
				copy(result.Before, recentLines[:len(recentLines)-1])
			}

			results = append(results, result)
		}
	}

	// Add after context (simplified - would need lookahead buffer)
	// For now, skipping after context to keep implementation simple

	return results
}

// formatOutput formats results based on output mode
func (t *GrepTool) formatOutput(results []searchResult, mode string) string {
	switch mode {
	case "files_with_matches":
		files := make(map[string]bool)
		for _, r := range results {
			relPath, _ := filepath.Rel(t.workDir, r.File)
			files[relPath] = true
		}

		var output strings.Builder
		for file := range files {
			output.WriteString(file + "\n")
		}
		return strings.TrimSpace(output.String())

	case "count":
		counts := make(map[string]int)
		for _, r := range results {
			relPath, _ := filepath.Rel(t.workDir, r.File)
			counts[relPath]++
		}

		var output strings.Builder
		for file, count := range counts {
			output.WriteString(fmt.Sprintf("%s: %d\n", file, count))
		}
		return strings.TrimSpace(output.String())

	case "content":
		var output strings.Builder
		currentFile := ""
		for _, r := range results {
			relPath, _ := filepath.Rel(t.workDir, r.File)
			if relPath != currentFile {
				if currentFile != "" {
					output.WriteString("\n")
				}
				output.WriteString(fmt.Sprintf("%s:\n", relPath))
				currentFile = relPath
			}

			// Add before context
			for _, line := range r.Before {
				output.WriteString(fmt.Sprintf("  %s\n", line))
			}

			// Add matching line
			output.WriteString(fmt.Sprintf("%d: %s\n", r.Line, r.Content))

			// Add after context
			for _, line := range r.After {
				output.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
		return strings.TrimSpace(output.String())

	default:
		return fmt.Sprintf("Found %d matches", len(results))
	}
}

// matchesFileType checks if extension matches file type
func matchesFileType(ext, fileType string) bool {
	typeMap := map[string][]string{
		"js":   {".js", ".jsx"},
		"ts":   {".ts", ".tsx"},
		"py":   {".py"},
		"go":   {".go"},
		"rust": {".rs"},
		"java": {".java"},
		"c":    {".c", ".h"},
		"cpp":  {".cpp", ".cc", ".cxx", ".hpp", ".hh"},
		"md":   {".md", ".markdown"},
	}

	exts, ok := typeMap[fileType]
	if !ok {
		return false
	}

	for _, e := range exts {
		if ext == e {
			return true
		}
	}
	return false
}
