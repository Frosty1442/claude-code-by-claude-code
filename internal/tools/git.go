package tools

import (
	"fmt"

	"github.com/claude-code-clone/claude-code-clone/internal/git"
	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// GitTool provides git operations
type GitTool struct {
	git *git.Git
}

// NewGitTool creates a new Git tool
func NewGitTool(workDir string) *GitTool {
	return &GitTool{
		git: git.NewGit(workDir),
	}
}

// Name returns the tool name
func (t *GitTool) Name() string {
	return "Git"
}

// Description returns the tool description
func (t *GitTool) Description() string {
	return "Perform git operations: status, diff, log, commit, push, pull, branch management"
}

// Schema returns the tool schema
func (t *GitTool) Schema() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        "Git",
		Description: "Perform git operations",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type": "string",
					"enum": []string{"status", "diff", "log", "add", "commit", "push", "pull", "branch", "smart-commit"},
					"description": "The git operation to perform",
				},
				"message": map[string]interface{}{
					"type":        "string",
					"description": "Commit message (for commit operation)",
				},
				"paths": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "File paths (for add operation)",
				},
				"remote": map[string]interface{}{
					"type":        "string",
					"description": "Remote name (for push/pull)",
					"default":     "origin",
				},
				"branch": map[string]interface{}{
					"type":        "string",
					"description": "Branch name",
				},
				"count": map[string]interface{}{
					"type":        "number",
					"description": "Number of log entries (for log operation)",
					"default":     10,
				},
				"force": map[string]interface{}{
					"type":        "boolean",
					"description": "Force push (dangerous, blocked for main/master)",
					"default":     false,
				},
			},
			"required": []string{"operation"},
		},
	}
}

// Validate validates the parameters
func (t *GitTool) Validate(params map[string]interface{}) error {
	operation, ok := getStringParam(params, "operation")
	if !ok {
		return fmt.Errorf("operation parameter is required")
	}

	validOps := map[string]bool{
		"status": true, "diff": true, "log": true, "add": true,
		"commit": true, "push": true, "pull": true, "branch": true, "smart-commit": true,
	}

	if !validOps[operation] {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	// Check if git repo
	if !t.git.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	return nil
}

// Execute runs the tool
func (t *GitTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	operation, _ := getStringParam(params, "operation")

	switch operation {
	case "status":
		return t.executeStatus()
	case "diff":
		return t.executeDiff()
	case "log":
		count, _ := getIntParam(params, "count")
		if count == 0 {
			count = 10
		}
		return t.executeLog(count)
	case "add":
		return t.executeAdd(params)
	case "commit":
		message, _ := getStringParam(params, "message")
		return t.executeCommit(message)
	case "smart-commit":
		return t.executeSmartCommit()
	case "push":
		return t.executePush(params)
	case "pull":
		return t.executePull(params)
	case "branch":
		return t.executeBranch(params)
	default:
		return NewErrorResult(fmt.Errorf("unknown operation: %s", operation)), nil
	}
}

func (t *GitTool) executeStatus() (*ToolResult, error) {
	output, err := t.git.Status()
	if err != nil {
		return NewErrorResult(err), nil
	}
	return NewSuccessResult(output), nil
}

func (t *GitTool) executeDiff() (*ToolResult, error) {
	output, err := t.git.Diff()
	if err != nil {
		return NewErrorResult(err), nil
	}
	return NewSuccessResult(output), nil
}

func (t *GitTool) executeLog(count int) (*ToolResult, error) {
	output, err := t.git.Log(count)
	if err != nil {
		return NewErrorResult(err), nil
	}
	return NewSuccessResult(output), nil
}

func (t *GitTool) executeAdd(params map[string]interface{}) (*ToolResult, error) {
	paths, ok := params["paths"]
	if !ok {
		return NewErrorResult(fmt.Errorf("paths parameter required for add operation")), nil
	}

	pathList, ok := paths.([]interface{})
	if !ok {
		return NewErrorResult(fmt.Errorf("paths must be an array")), nil
	}

	pathStrs := make([]string, len(pathList))
	for i, p := range pathList {
		pathStrs[i] = fmt.Sprintf("%v", p)
	}

	err := t.git.Add(pathStrs...)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(fmt.Sprintf("Added %d file(s) to staging", len(pathStrs))), nil
}

func (t *GitTool) executeCommit(message string) (*ToolResult, error) {
	if message == "" {
		return NewErrorResult(fmt.Errorf("commit message required")), nil
	}

	// Check for secrets
	secrets, _ := t.git.DetectSecretsInStaged()
	if len(secrets) > 0 {
		return NewErrorResult(fmt.Errorf("potential secrets detected: %v. Please review before committing", secrets)), nil
	}

	err := t.git.Commit(message)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(fmt.Sprintf("Created commit: %s", message)), nil
}

func (t *GitTool) executeSmartCommit() (*ToolResult, error) {
	// Check for secrets
	secrets, _ := t.git.DetectSecretsInStaged()
	if len(secrets) > 0 {
		return NewErrorResult(fmt.Errorf("potential secrets detected: %v. Please review before committing", secrets)), nil
	}

	message, err := t.git.SmartCommit()
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(fmt.Sprintf("Created commit with generated message: %s", message)), nil
}

func (t *GitTool) executePush(params map[string]interface{}) (*ToolResult, error) {
	remote, ok := getStringParam(params, "remote")
	if !ok {
		remote = "origin"
	}

	branch, ok := getStringParam(params, "branch")
	if !ok {
		// Get current branch
		var err error
		branch, err = t.git.CurrentBranch()
		if err != nil {
			return NewErrorResult(err), nil
		}
	}

	force, _ := getBoolParam(params, "force")

	err := t.git.Push(remote, branch, force)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(fmt.Sprintf("Pushed to %s/%s", remote, branch)), nil
}

func (t *GitTool) executePull(params map[string]interface{}) (*ToolResult, error) {
	remote, ok := getStringParam(params, "remote")
	if !ok {
		remote = "origin"
	}

	branch, ok := getStringParam(params, "branch")
	if !ok {
		var err error
		branch, err = t.git.CurrentBranch()
		if err != nil {
			return NewErrorResult(err), nil
		}
	}

	err := t.git.Pull(remote, branch)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(fmt.Sprintf("Pulled from %s/%s", remote, branch)), nil
}

func (t *GitTool) executeBranch(params map[string]interface{}) (*ToolResult, error) {
	branchName, ok := getStringParam(params, "branch")
	if !ok {
		// Get current branch
		current, err := t.git.CurrentBranch()
		if err != nil {
			return NewErrorResult(err), nil
		}
		return NewSuccessResult(fmt.Sprintf("Current branch: %s", current)), nil
	}

	// Create or switch to branch
	err := t.git.SwitchBranch(branchName)
	if err != nil {
		// Try creating it
		err = t.git.CreateBranch(branchName)
		if err != nil {
			return NewErrorResult(err), nil
		}
		return NewSuccessResult(fmt.Sprintf("Created and switched to branch: %s", branchName)), nil
	}

	return NewSuccessResult(fmt.Sprintf("Switched to branch: %s", branchName)), nil
}
