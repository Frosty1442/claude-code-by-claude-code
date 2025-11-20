package tools

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebFetchTool_Name(t *testing.T) {
	tool := NewWebFetchTool()
	if tool.Name() != "WebFetch" {
		t.Errorf("Expected name 'WebFetch', got '%s'", tool.Name())
	}
}

func TestWebFetchTool_Execute(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Page</h1><p>Test content</p></body></html>"))
	}))
	defer server.Close()

	tool := NewWebFetchTool()

	params := map[string]interface{}{
		"url": server.URL,
	}

	result, err := tool.Execute(params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Type != "success" {
		t.Errorf("Expected success result, got %s", result.Type)
	}

	if result.Output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestWebFetchTool_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test content"))
	}))
	defer server.Close()

	tool := NewWebFetchTool()

	params := map[string]interface{}{
		"url": server.URL,
	}

	// First call
	tool.Execute(params)

	// Second call should use cache
	tool.Execute(params)

	// Should only call server once due to caching
	if callCount != 1 {
		t.Errorf("Expected server to be called once (cached), got %d calls", callCount)
	}

	// Test with cache disabled
	params["use_cache"] = false
	tool.Execute(params)

	if callCount != 2 {
		t.Errorf("Expected server to be called twice (no cache), got %d calls", callCount)
	}
}

func TestWebFetchTool_InvalidURL(t *testing.T) {
	tool := NewWebFetchTool()

	params := map[string]interface{}{
		"url": "not-a-valid-url",
	}

	result, _ := tool.Execute(params)
	if result.Type != "error" {
		t.Errorf("Expected error result for invalid URL, got %s", result.Type)
	}
}

func TestWebFetchTool_HTMLToText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
			<head><title>Test</title></head>
			<body>
				<h1>Heading</h1>
				<p>Paragraph text</p>
				<script>alert('should be removed')</script>
				<style>body { color: red; }</style>
			</body>
			</html>
		`))
	}))
	defer server.Close()

	tool := NewWebFetchTool()

	params := map[string]interface{}{
		"url": server.URL,
	}

	result, err := tool.Execute(params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should contain text content
	if !contains(result.Output, "Heading") || !contains(result.Output, "Paragraph text") {
		t.Error("Expected text content from HTML")
	}

	// Should not contain script or style tags
	if contains(result.Output, "alert") || contains(result.Output, "color: red") {
		t.Error("Script and style tags should be removed")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
