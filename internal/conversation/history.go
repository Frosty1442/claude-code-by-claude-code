package conversation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// HistoryManager manages conversation history persistence
type HistoryManager struct {
	historyDir string
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(workDir string) *HistoryManager {
	historyDir := filepath.Join(workDir, ".claude", "history")
	os.MkdirAll(historyDir, 0755)

	return &HistoryManager{
		historyDir: historyDir,
	}
}

// SaveHistory saves conversation history to disk
func (h *HistoryManager) SaveHistory(sessionID string, messages []schema.Message) error {
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().Unix())
	}

	filePath := filepath.Join(h.historyDir, sessionID+".json")

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// LoadHistory loads conversation history from disk
func (h *HistoryManager) LoadHistory(sessionID string) ([]schema.Message, error) {
	filePath := filepath.Join(h.historyDir, sessionID+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history: %w", err)
	}

	var messages []schema.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return messages, nil
}

// ListSessions lists available session IDs
func (h *HistoryManager) ListSessions() ([]string, error) {
	entries, err := os.ReadDir(h.historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	sessions := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			sessionID := entry.Name()[:len(entry.Name())-5] // Remove .json
			sessions = append(sessions, sessionID)
		}
	}

	return sessions, nil
}

// DeleteHistory deletes a session history
func (h *HistoryManager) DeleteHistory(sessionID string) error {
	filePath := filepath.Join(h.historyDir, sessionID+".json")
	return os.Remove(filePath)
}
