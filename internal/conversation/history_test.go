package conversation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

func TestNewHistoryManager(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := NewHistoryManager(tmpDir)
	if mgr == nil {
		t.Fatal("NewHistoryManager returned nil")
	}

	// Check that .claude/history directory was created
	historyDir := filepath.Join(tmpDir, ".claude", "history")
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		t.Error("Expected .claude/history directory to be created")
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewHistoryManager(tmpDir)

	sessionID := "test-session-123"

	// Create test messages
	messages := []schema.Message{
		{
			Role: "user",
			Content: []schema.ContentBlock{
				{Type: "text", Text: "Hello"},
			},
		},
		{
			Role: "assistant",
			Content: []schema.ContentBlock{
				{Type: "text", Text: "Hi there!"},
			},
		},
		{
			Role: "user",
			Content: []schema.ContentBlock{
				{Type: "text", Text: "How are you?"},
			},
		},
	}

	// Save history
	err := mgr.SaveHistory(sessionID, messages)
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	// Load history
	loadedMessages, err := mgr.LoadHistory(sessionID)
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	// Verify loaded messages match saved messages
	if len(loadedMessages) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(loadedMessages))
	}

	for i, msg := range loadedMessages {
		if msg.Role != messages[i].Role {
			t.Errorf("Message %d: expected role '%s', got '%s'", i, messages[i].Role, msg.Role)
		}

		if len(msg.Content) != len(messages[i].Content) {
			t.Errorf("Message %d: expected %d content blocks, got %d", i, len(messages[i].Content), len(msg.Content))
		}
	}
}

func TestSaveHistory_AutoGenerateSessionID(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewHistoryManager(tmpDir)

	messages := []schema.Message{
		{
			Role: "user",
			Content: []schema.ContentBlock{
				{Type: "text", Text: "Test message"},
			},
		},
	}

	// Save with empty session ID (should auto-generate)
	err := mgr.SaveHistory("", messages)
	if err != nil {
		t.Fatalf("SaveHistory with empty ID failed: %v", err)
	}

	// List sessions to verify it was saved
	sessions, err := mgr.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) == 0 {
		t.Error("Expected at least one session after saving with auto-generated ID")
	}
}

func TestLoadHistory_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewHistoryManager(tmpDir)

	// Try to load non-existent session
	_, err := mgr.LoadHistory("non-existent-session")
	if err == nil {
		t.Error("Expected error when loading non-existent session")
	}
}

func TestListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewHistoryManager(tmpDir)

	// Create multiple sessions
	sessions := []string{"session-1", "session-2", "session-3"}
	messages := []schema.Message{
		{
			Role: "user",
			Content: []schema.ContentBlock{
				{Type: "text", Text: "Test"},
			},
		},
	}

	for _, sessionID := range sessions {
		err := mgr.SaveHistory(sessionID, messages)
		if err != nil {
			t.Fatalf("SaveHistory failed for %s: %v", sessionID, err)
		}
	}

	// List sessions
	listedSessions, err := mgr.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(listedSessions) != len(sessions) {
		t.Errorf("Expected %d sessions, got %d", len(sessions), len(listedSessions))
	}

	// Verify all sessions are in the list
	for _, sessionID := range sessions {
		found := false
		for _, listed := range listedSessions {
			if listed == sessionID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Session %s not found in list", sessionID)
		}
	}
}

func TestListSessions_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewHistoryManager(tmpDir)

	// List sessions in empty directory
	sessions, err := mgr.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions in empty directory, got %d", len(sessions))
	}
}

func TestDeleteHistory(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewHistoryManager(tmpDir)

	sessionID := "session-to-delete"
	messages := []schema.Message{
		{
			Role: "user",
			Content: []schema.ContentBlock{
				{Type: "text", Text: "Test"},
			},
		},
	}

	// Save session
	err := mgr.SaveHistory(sessionID, messages)
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	// Verify it exists
	sessions, _ := mgr.ListSessions()
	if len(sessions) == 0 {
		t.Fatal("Session was not saved")
	}

	// Delete session
	err = mgr.DeleteHistory(sessionID)
	if err != nil {
		t.Fatalf("DeleteHistory failed: %v", err)
	}

	// Verify it's gone
	sessions, _ = mgr.ListSessions()
	for _, s := range sessions {
		if s == sessionID {
			t.Error("Session should have been deleted")
		}
	}
}

func TestDeleteHistory_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewHistoryManager(tmpDir)

	// Try to delete non-existent session
	err := mgr.DeleteHistory("non-existent-session")
	if err == nil {
		t.Error("Expected error when deleting non-existent session")
	}
}
