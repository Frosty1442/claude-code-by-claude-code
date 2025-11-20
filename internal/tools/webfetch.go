package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// WebFetchTool fetches content from URLs
type WebFetchTool struct {
	cache      map[string]*cacheEntry
	httpClient *http.Client
}

type cacheEntry struct {
	content   string
	timestamp time.Time
}

// NewWebFetchTool creates a new WebFetch tool
func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		cache: make(map[string]*cacheEntry),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
}

// Name returns the tool name
func (t *WebFetchTool) Name() string {
	return "WebFetch"
}

// Description returns the tool description
func (t *WebFetchTool) Description() string {
	return "Fetch content from a URL"
}

// Schema returns the tool schema
func (t *WebFetchTool) Schema() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        "WebFetch",
		Description: "Fetch content from a URL (supports caching)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The URL to fetch",
				},
				"use_cache": map[string]interface{}{
					"type":        "boolean",
					"description": "Use cached content if available (15 min TTL)",
					"default":     true,
				},
			},
			"required": []string{"url"},
		},
	}
}

// Validate validates the parameters
func (t *WebFetchTool) Validate(params map[string]interface{}) error {
	urlStr, ok := getStringParam(params, "url")
	if !ok {
		return fmt.Errorf("url parameter is required")
	}

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS URLs are supported")
	}

	return nil
}

// Execute runs the tool
func (t *WebFetchTool) Execute(params map[string]interface{}) (*ToolResult, error) {
	urlStr, _ := getStringParam(params, "url")
	useCache, ok := getBoolParam(params, "use_cache")
	if !ok {
		useCache = true
	}

	// Upgrade http to https (except for localhost/127.0.0.1 for testing)
	if strings.HasPrefix(urlStr, "http://") &&
		!strings.Contains(urlStr, "localhost") &&
		!strings.Contains(urlStr, "127.0.0.1") {
		urlStr = "https://" + strings.TrimPrefix(urlStr, "http://")
	}

	// Check cache
	if useCache {
		if entry, ok := t.cache[urlStr]; ok {
			// Check if cache is still valid (15 minutes)
			if time.Since(entry.timestamp) < 15*time.Minute {
				return NewSuccessResult(fmt.Sprintf("[CACHED]\n\n%s", entry.content)), nil
			}
		}
	}

	// Fetch URL
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return NewErrorResult(err), nil
	}

	req.Header.Set("User-Agent", "Claude-Code-Clone/1.0")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return NewErrorResult(err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewErrorResult(fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)), nil
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewErrorResult(err), nil
	}

	content := string(body)

	// Convert HTML to plain text (simple version)
	if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		content = t.htmlToText(content)
	}

	// Truncate if too long
	if len(content) > 50000 {
		content = content[:50000] + "\n\n[Content truncated - too large]"
	}

	// Cache the result
	t.cache[urlStr] = &cacheEntry{
		content:   content,
		timestamp: time.Now(),
	}

	return NewSuccessResult(content), nil
}

// htmlToText converts HTML to plain text (very basic)
func (t *WebFetchTool) htmlToText(html string) string {
	// Remove script and style tags
	html = removeTag(html, "script")
	html = removeTag(html, "style")

	// Remove HTML tags (keep content)
	text := strings.Builder{}
	inTag := false

	for _, char := range html {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			text.WriteRune(' ')
			continue
		}
		if !inTag {
			text.WriteRune(char)
		}
	}

	// Clean up whitespace
	result := text.String()
	result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	result = strings.TrimSpace(result)

	return result
}

// removeTag removes HTML tags and their content
func removeTag(html, tag string) string {
	startTag := "<" + tag
	endTag := "</" + tag + ">"

	for {
		start := strings.Index(strings.ToLower(html), startTag)
		if start == -1 {
			break
		}

		end := strings.Index(strings.ToLower(html[start:]), endTag)
		if end == -1 {
			break
		}

		html = html[:start] + html[start+end+len(endTag):]
	}

	return html
}
