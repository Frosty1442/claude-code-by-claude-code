package tools

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// BackgroundShellManager manages background shell processes
type BackgroundShellManager struct {
	shells map[string]*BackgroundShell
	mu     sync.RWMutex
	nextID int
}

// BackgroundShell represents a running background shell
type BackgroundShell struct {
	ID      string
	Command string
	Cmd     *exec.Cmd
	Stdout  *bufio.Scanner
	Stderr  *bufio.Scanner
	Done    chan bool
	Output  []string
	mu      sync.Mutex
}

// NewBackgroundShellManager creates a new manager
func NewBackgroundShellManager() *BackgroundShellManager {
	return &BackgroundShellManager{
		shells: make(map[string]*BackgroundShell),
		nextID: 1,
	}
}

// Start starts a command in the background
func (m *BackgroundShellManager) Start(workDir, command string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID
	id := fmt.Sprintf("bg-%d", m.nextID)
	m.nextID++

	// Create command
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = workDir

	// Setup pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	// Create shell
	shell := &BackgroundShell{
		ID:      id,
		Command: command,
		Cmd:     cmd,
		Stdout:  bufio.NewScanner(stdout),
		Stderr:  bufio.NewScanner(stderr),
		Done:    make(chan bool),
		Output:  []string{},
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Start output readers
	go shell.readOutput()

	// Wait for completion in background
	go func() {
		cmd.Wait()
		close(shell.Done)
	}()

	m.shells[id] = shell
	return id, nil
}

// GetOutput gets new output from a background shell
func (m *BackgroundShellManager) GetOutput(id string) ([]string, bool, error) {
	m.mu.RLock()
	shell, ok := m.shells[id]
	m.mu.RUnlock()

	if !ok {
		return nil, false, fmt.Errorf("shell not found: %s", id)
	}

	shell.mu.Lock()
	output := make([]string, len(shell.Output))
	copy(output, shell.Output)
	shell.Output = []string{} // Clear after reading
	shell.mu.Unlock()

	select {
	case <-shell.Done:
		return output, true, nil // Command finished
	default:
		return output, false, nil // Still running
	}
}

// Kill kills a background shell
func (m *BackgroundShellManager) Kill(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	shell, ok := m.shells[id]
	if !ok {
		return fmt.Errorf("shell not found: %s", id)
	}

	if shell.Cmd.Process != nil {
		shell.Cmd.Process.Kill()
	}

	delete(m.shells, id)
	return nil
}

// List lists all background shells
func (m *BackgroundShellManager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.shells))
	for id := range m.shells {
		ids = append(ids, id)
	}
	return ids
}

// readOutput reads output from stdout and stderr
func (s *BackgroundShell) readOutput() {
	go func() {
		for s.Stdout.Scan() {
			s.mu.Lock()
			s.Output = append(s.Output, s.Stdout.Text())
			s.mu.Unlock()
		}
	}()

	go func() {
		for s.Stderr.Scan() {
			s.mu.Lock()
			s.Output = append(s.Output, "[STDERR] "+s.Stderr.Text())
			s.mu.Unlock()
		}
	}()
}

// Global background shell manager
var globalShellManager = NewBackgroundShellManager()

// GetGlobalShellManager returns the global shell manager
func GetGlobalShellManager() *BackgroundShellManager {
	return globalShellManager
}

// BashOutputTool gets output from background shell
type BashOutputTool struct{}

func NewBashOutputTool() *BashOutputTool {
	return &BashOutputTool{}
}

func (t *BashOutputTool) Name() string {
	return "BashOutput"
}

func (t *BashOutputTool) Description() string {
	return "Get output from a background shell"
}

func (t *BashOutputTool) Schema() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        "BashOutput",
		Description: "Get output from a background shell",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"bash_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the background shell",
				},
			},
			"required": []string{"bash_id"},
		},
	}
}

func (t *BashOutputTool) Validate(params map[string]interface{}) error {
	if _, ok := getStringParam(params, "bash_id"); !ok {
		return fmt.Errorf("bash_id parameter is required")
	}
	return nil
}

func (t *BashOutputTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	bashID, _ := getStringParam(params, "bash_id")

	output, done, err := GetGlobalShellManager().GetOutput(bashID)
	if err != nil {
		return NewErrorResult(err), nil
	}

	result := fmt.Sprintf("Shell %s: %d new lines\n", bashID, len(output))
	for _, line := range output {
		result += line + "\n"
	}

	if done {
		result += "\n[Shell completed]"
	} else {
		result += "\n[Shell still running]"
	}

	return NewSuccessResult(result), nil
}

// KillShellTool kills a background shell
type KillShellTool struct{}

func NewKillShellTool() *KillShellTool {
	return &KillShellTool{}
}

func (t *KillShellTool) Name() string {
	return "KillShell"
}

func (t *KillShellTool) Description() string {
	return "Kill a background shell"
}

func (t *KillShellTool) Schema() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        "KillShell",
		Description: "Kill a background shell",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"shell_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the shell to kill",
				},
			},
			"required": []string{"shell_id"},
		},
	}
}

func (t *KillShellTool) Validate(params map[string]interface{}) error {
	if _, ok := getStringParam(params, "shell_id"); !ok {
		return fmt.Errorf("shell_id parameter is required")
	}
	return nil
}

func (t *KillShellTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	shellID, _ := getStringParam(params, "shell_id")

	err := GetGlobalShellManager().Kill(shellID)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(fmt.Sprintf("Killed shell: %s", shellID)), nil
}
