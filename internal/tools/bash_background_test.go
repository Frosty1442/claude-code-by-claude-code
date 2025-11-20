package tools

import (
	"testing"
	"time"
)

func TestBackgroundShellManager_Start(t *testing.T) {
	mgr := NewBackgroundShellManager()

	id, err := mgr.Start("/tmp", "echo 'Hello World'")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if id == "" {
		t.Error("Expected non-empty shell ID")
	}

	// Wait for command to complete
	time.Sleep(100 * time.Millisecond)

	// Check output
	output, done, err := mgr.GetOutput(id)
	if err != nil {
		t.Fatalf("GetOutput failed: %v", err)
	}

	if !done {
		t.Error("Expected command to be done")
	}

	if len(output) == 0 {
		t.Error("Expected output from command")
	}
}

func TestBackgroundShellManager_Kill(t *testing.T) {
	mgr := NewBackgroundShellManager()

	// Start a long-running command
	id, err := mgr.Start("/tmp", "sleep 10")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Kill it
	err = mgr.Kill(id)
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}

	// Verify it's gone
	_, _, err = mgr.GetOutput(id)
	if err == nil {
		t.Error("Expected error when getting output from killed shell")
	}
}

func TestBackgroundShellManager_List(t *testing.T) {
	mgr := NewBackgroundShellManager()

	// Start multiple commands
	id1, _ := mgr.Start("/tmp", "sleep 1")
	id2, _ := mgr.Start("/tmp", "sleep 1")

	// List shells
	ids := mgr.List()

	if len(ids) < 2 {
		t.Errorf("Expected at least 2 shells, got %d", len(ids))
	}

	// Check that our IDs are in the list
	found1, found2 := false, false
	for _, id := range ids {
		if id == id1 {
			found1 = true
		}
		if id == id2 {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Expected both shell IDs in list")
	}

	// Cleanup
	mgr.Kill(id1)
	mgr.Kill(id2)
}

func TestBashOutputTool(t *testing.T) {
	mgr := GetGlobalShellManager() // Use global manager
	tool := NewBashOutputTool()

	// Start a command
	id, err := mgr.Start("/tmp", "echo 'test output'")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for output
	time.Sleep(100 * time.Millisecond)

	// Get output using tool
	params := map[string]interface{}{
		"bash_id": id,
	}

	result, err := tool.Execute(params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type != "success" {
		t.Errorf("Expected success result, got %s: %s", result.Type, result.Output)
	}

	if result.Output == "" {
		t.Error("Expected non-empty output")
	}

	// Cleanup
	mgr.Kill(id)
}

func TestKillShellTool(t *testing.T) {
	mgr := GetGlobalShellManager() // Use global manager
	tool := NewKillShellTool()

	// Start a command
	id, err := mgr.Start("/tmp", "sleep 10")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Kill using tool
	params := map[string]interface{}{
		"shell_id": id,
	}

	result, err := tool.Execute(params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type != "success" {
		t.Errorf("Expected success result, got %s: %s", result.Type, result.Output)
	}

	// Verify shell is gone
	_, _, err = mgr.GetOutput(id)
	if err == nil {
		t.Error("Expected error after killing shell")
	}
}
