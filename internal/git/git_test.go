package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGit(t *testing.T) {
	tmpDir := t.TempDir()
	g := NewGit(tmpDir)

	if g == nil {
		t.Fatal("NewGit returned nil")
	}

	if g.workDir != tmpDir {
		t.Errorf("Expected workDir %s, got %s", tmpDir, g.workDir)
	}
}

func TestStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	g := NewGit(tmpDir)
	if _, err := g.runCommand("init"); err != nil {
		t.Skipf("Git not available: %v", err)
	}

	// Test status on clean repo
	status, err := g.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if status == "" {
		t.Error("Expected non-empty status")
	}
}

func TestDiff(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	g := NewGit(tmpDir)
	if _, err := g.runCommand("init"); err != nil {
		t.Skipf("Git not available: %v", err)
	}

	// Configure git
	g.runCommand("config", "user.name", "Test User")
	g.runCommand("config", "user.email", "test@example.com")

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("original"), 0644)
	g.runCommand("add", "test.txt")
	g.runCommand("commit", "-m", "Initial commit")

	// Modify file
	os.WriteFile(testFile, []byte("modified"), 0644)

	// Test diff
	diff, err := g.Diff()
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if diff == "" {
		t.Error("Expected non-empty diff")
	}
}

func TestDetectSecretsInStaged(t *testing.T) {
	tmpDir := t.TempDir()

	g := NewGit(tmpDir)
	if _, err := g.runCommand("init"); err != nil {
		t.Skipf("Git not available: %v", err)
	}

	// Configure git
	g.runCommand("config", "user.name", "Test User")
	g.runCommand("config", "user.email", "test@example.com")

	tests := []struct {
		name     string
		filename string
		content  string
		hasSecret bool
	}{
		{
			name:     "API key",
			filename: "config.txt",
			content:  "API_KEY=sk-1234567890abcdef",
			hasSecret: true,
		},
		{
			name:     "Password",
			filename: "db.conf",
			content:  "password=mysecretpass123",
			hasSecret: true,
		},
		{
			name:     "Normal file",
			filename: "readme.md",
			content:  "# README\n\nThis is a normal file",
			hasSecret: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create file
			filePath := filepath.Join(tmpDir, tt.filename)
			os.WriteFile(filePath, []byte(tt.content), 0644)

			// Stage file
			g.runCommand("add", tt.filename)

			// Detect secrets
			secrets, err := g.DetectSecretsInStaged()
			if err != nil {
				t.Fatalf("DetectSecretsInStaged failed: %v", err)
			}

			if tt.hasSecret && len(secrets) == 0 {
				t.Errorf("Expected secrets to be detected in %s", tt.filename)
			}

			// Clean up for next test
			g.runCommand("reset", "HEAD", tt.filename)
			os.Remove(filePath)
		})
	}
}

func TestCommit(t *testing.T) {
	tmpDir := t.TempDir()

	g := NewGit(tmpDir)
	if _, err := g.runCommand("init"); err != nil {
		t.Skipf("Git not available: %v", err)
	}

	// Configure git
	g.runCommand("config", "user.name", "Test User")
	g.runCommand("config", "user.email", "test@example.com")

	// Create and stage a file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)
	g.runCommand("add", "test.txt")

	// Test commit
	err := g.Commit("Test commit message")
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify commit exists
	log, err := g.Log(1)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	if log == "" {
		t.Error("Expected non-empty log")
	}
}
