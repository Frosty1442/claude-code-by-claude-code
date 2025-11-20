package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Git provides git operations
type Git struct {
	workDir string
}

// NewGit creates a new Git instance
func NewGit(workDir string) *Git {
	return &Git{workDir: workDir}
}

// Status returns git status
func (g *Git) Status() (string, error) {
	return g.runCommand("git", "status")
}

// Diff returns git diff (both staged and unstaged)
func (g *Git) Diff() (string, error) {
	// Get unstaged changes
	unstaged, _ := g.runCommand("git", "diff")

	// Get staged changes
	staged, _ := g.runCommand("git", "diff", "--cached")

	result := ""
	if unstaged != "" {
		result += "=== Unstaged Changes ===\n" + unstaged + "\n"
	}
	if staged != "" {
		result += "=== Staged Changes ===\n" + staged + "\n"
	}

	if result == "" {
		return "No changes", nil
	}

	return result, nil
}

// Log returns recent commit log
func (g *Git) Log(count int) (string, error) {
	if count <= 0 {
		count = 10
	}
	return g.runCommand("git", "log", fmt.Sprintf("-%d", count), "--oneline")
}

// Add adds files to staging
func (g *Git) Add(paths ...string) error {
	args := append([]string{"add"}, paths...)
	_, err := g.runCommand("git", args...)
	return err
}

// Commit creates a commit
func (g *Git) Commit(message string) error {
	_, err := g.runCommand("git", "commit", "-m", message)
	return err
}

// SmartCommit analyzes changes and creates a commit with generated message
func (g *Git) SmartCommit() (string, error) {
	// Get diff to analyze
	diff, err := g.Diff()
	if err != nil {
		return "", err
	}

	if diff == "No changes" {
		return "", fmt.Errorf("no changes to commit")
	}

	// Get recent commits for style
	recentLog, _ := g.Log(5)

	// Generate commit message (simplified - in production would use LLM)
	message := g.generateCommitMessage(diff, recentLog)

	// Create commit
	err = g.Commit(message)
	if err != nil {
		return "", err
	}

	return message, nil
}

// generateCommitMessage generates a commit message from diff
func (g *Git) generateCommitMessage(diff, recentLog string) string {
	// Simple heuristic-based message generation
	// In production, would use LLM to analyze diff

	lines := strings.Split(diff, "\n")

	// Count additions and deletions
	additions := 0
	deletions := 0
	files := make(map[string]bool)

	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			additions++
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			deletions++
		}
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			// Extract filename
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				file := filepath.Base(parts[1])
				files[file] = true
			}
		}
	}

	// Determine change type
	changeType := "update"
	if additions > deletions*2 {
		changeType = "feat"
	} else if deletions > additions*2 {
		changeType = "refactor"
	}

	// Build message
	fileList := ""
	if len(files) <= 3 {
		fileNames := make([]string, 0, len(files))
		for file := range files {
			fileNames = append(fileNames, file)
		}
		fileList = strings.Join(fileNames, ", ")
	} else {
		fileList = fmt.Sprintf("%d files", len(files))
	}

	return fmt.Sprintf("%s: Update %s", changeType, fileList)
}

// Push pushes to remote
func (g *Git) Push(remote, branch string, force bool) error {
	args := []string{"push"}

	if force {
		// Safety check for main/master
		if branch == "main" || branch == "master" {
			return fmt.Errorf("force push to %s is not allowed for safety", branch)
		}
		args = append(args, "--force")
	}

	args = append(args, "-u", remote, branch)

	_, err := g.runCommandWithRetry("git", args...)
	return err
}

// Pull pulls from remote
func (g *Git) Pull(remote, branch string) error {
	_, err := g.runCommandWithRetry("git", "pull", remote, branch)
	return err
}

// CreateBranch creates a new branch
func (g *Git) CreateBranch(name string) error {
	_, err := g.runCommand("git", "checkout", "-b", name)
	return err
}

// SwitchBranch switches to a branch
func (g *Git) SwitchBranch(name string) error {
	_, err := g.runCommand("git", "checkout", name)
	return err
}

// CurrentBranch returns the current branch name
func (g *Git) CurrentBranch() (string, error) {
	output, err := g.runCommand("git", "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// IsGitRepo checks if the directory is a git repository
func (g *Git) IsGitRepo() bool {
	_, err := g.runCommand("git", "rev-parse", "--git-dir")
	return err == nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func (g *Git) HasUncommittedChanges() bool {
	output, err := g.runCommand("git", "status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) != ""
}

// DetectSecretsInStaged checks staged files for potential secrets
func (g *Git) DetectSecretsInStaged() ([]string, error) {
	diff, err := g.runCommand("git", "diff", "--cached")
	if err != nil {
		return nil, err
	}

	secrets := []string{}

	// Patterns for common secrets
	patterns := []struct {
		pattern *regexp.Regexp
		name    string
	}{
		{regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]?[a-zA-Z0-9]{20,}['"]?`), "API Key"},
		{regexp.MustCompile(`(?i)(secret|password|passwd)\s*[:=]\s*['"]?[^\s'"]{8,}['"]?`), "Password/Secret"},
		{regexp.MustCompile(`(?i)token\s*[:=]\s*['"]?[a-zA-Z0-9]{20,}['"]?`), "Token"},
		{regexp.MustCompile(`(?i)(aws|amazon).*(key|secret)`), "AWS Credentials"},
		{regexp.MustCompile(`-----BEGIN (RSA|OPENSSH|DSA|EC) PRIVATE KEY-----`), "Private Key"},
	}

	for _, p := range patterns {
		if p.pattern.MatchString(diff) {
			secrets = append(secrets, p.name)
		}
	}

	return secrets, nil
}

// runCommand executes a git command
func (g *Git) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = g.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%s: %s", err, stderr.String())
		}
		return "", err
	}

	return stdout.String(), nil
}

// runCommandWithRetry executes a command with exponential backoff retry
func (g *Git) runCommandWithRetry(name string, args ...string) (string, error) {
	maxRetries := 4
	backoff := []int{2, 4, 8, 16} // seconds

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		output, err := g.runCommand(name, args...)
		if err == nil {
			return output, nil
		}

		lastErr = err

		// Check if it's a network error
		if !isNetworkError(err) {
			return "", err
		}

		if attempt < maxRetries {
			// Wait before retry
			exec.Command("sleep", fmt.Sprintf("%d", backoff[attempt])).Run()
		}
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// isNetworkError checks if an error is network-related
func isNetworkError(err error) bool {
	errStr := err.Error()
	networkErrors := []string{
		"network",
		"timeout",
		"connection refused",
		"could not resolve",
		"temporary failure",
	}

	for _, ne := range networkErrors {
		if strings.Contains(strings.ToLower(errStr), ne) {
			return true
		}
	}

	return false
}
