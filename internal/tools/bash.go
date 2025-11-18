package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// BashTool executes bash commands
type BashTool struct {
	workDir string
}

// NewBashTool creates a new Bash tool
func NewBashTool(workDir string) *BashTool {
	return &BashTool{workDir: workDir}
}

// Name returns the tool name
func (t *BashTool) Name() string {
	return "Bash"
}

// Description returns the tool description
func (t *BashTool) Description() string {
	return "Executes a bash command with optional timeout"
}

// Schema returns the tool schema
func (t *BashTool) Schema() schema.ToolDefinition {
	return schema.BashToolSchema
}

// Validate validates the parameters
func (t *BashTool) Validate(params map[string]interface{}) error {
	if _, ok := getStringParam(params, "command"); !ok {
		return fmt.Errorf("command parameter is required")
	}
	return nil
}

// Execute runs the tool
func (t *BashTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	command, _ := getStringParam(params, "command")
	timeoutMs, ok := getIntParam(params, "timeout")
	if !ok {
		timeoutMs = 120000 // 2 minutes default
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// Execute command
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = t.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Prepare output
	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return NewErrorResult(fmt.Errorf("command timed out after %dms", timeoutMs)), nil
		}
		// Command failed but still return output
		if output == "" {
			output = err.Error()
		}
		return NewSystemResult(output), nil
	}

	if output == "" {
		return NewSystemResult("Tool ran without output or errors"), nil
	}

	return NewSuccessResult(output), nil
}
