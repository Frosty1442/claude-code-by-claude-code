package conversation

import (
	"encoding/json"
	"regexp"
	"strings"
)

// CompatibilityMode enables tool use with models that don't support native function calling
// by using structured output parsing (ReAct pattern, XML tags, or JSON blocks)

// ParseToolRequestFromText extracts tool requests from model text output
// Supports multiple formats:
// 1. XML tags: <tool_use name="Read"><file_path>/path/to/file</file_path></tool_use>
// 2. JSON blocks: ```json\n{"tool": "Read", "params": {"file_path": "/path"}}\n```
// 3. ReAct format: Action: Read[file_path="/path/to/file"]
func ParseToolRequestFromText(text string) []ToolRequest {
	var requests []ToolRequest

	// Try XML format first
	requests = append(requests, parseXMLToolRequests(text)...)

	// Try JSON code blocks
	requests = append(requests, parseJSONToolRequests(text)...)

	// Try ReAct format
	requests = append(requests, parseReActToolRequests(text)...)

	return requests
}

// ToolRequest represents a parsed tool request from text
type ToolRequest struct {
	Name       string
	Parameters map[string]interface{}
}

// parseXMLToolRequests extracts tool calls from XML-like tags
// Example: <tool_use name="Read"><file_path>/home/user/file.txt</file_path></tool_use>
func parseXMLToolRequests(text string) []ToolRequest {
	var requests []ToolRequest

	// Match <tool_use name="ToolName">...</tool_use> (non-greedy)
	toolPattern := regexp.MustCompile(`(?s)<tool_use\s+name="([^"]+)">(.*?)</tool_use>`)
	matches := toolPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			toolName := match[1]
			paramsXML := match[2]

			params := parseXMLParams(paramsXML)
			requests = append(requests, ToolRequest{
				Name:       toolName,
				Parameters: params,
			})
		}
	}

	return requests
}

// parseXMLParams extracts parameters from XML content
// Example: <file_path>/path</file_path><limit>10</limit>
func parseXMLParams(xml string) map[string]interface{} {
	params := make(map[string]interface{})

	// Match <tag>value</tag>  - note: Go regexp doesn't support backreferences
	// So we match any <tag>value</tag> pattern
	paramPattern := regexp.MustCompile(`<([^>]+)>([^<]*)</[^>]+>`)
	matches := paramPattern.FindAllStringSubmatch(xml, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			paramName := match[1]
			paramValue := match[2]
			params[paramName] = paramValue
		}
	}

	return params
}

// parseJSONToolRequests extracts tool calls from JSON code blocks
// Example: ```json\n{"tool": "Read", "params": {"file_path": "/path"}}\n```
func parseJSONToolRequests(text string) []ToolRequest {
	var requests []ToolRequest

	// Match ```json ... ```
	jsonPattern := regexp.MustCompile("(?s)```json\\s*\\n(.*?)\\n```")
	matches := jsonPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			var toolReq struct {
				Tool   string                 `json:"tool"`
				Params map[string]interface{} `json:"params"`
			}

			if err := json.Unmarshal([]byte(match[1]), &toolReq); err == nil {
				requests = append(requests, ToolRequest{
					Name:       toolReq.Tool,
					Parameters: toolReq.Params,
				})
			}
		}
	}

	return requests
}

// parseReActToolRequests extracts tool calls from ReAct format
// Example: Action: Read[file_path="/home/user/file.txt"]
func parseReActToolRequests(text string) []ToolRequest {
	var requests []ToolRequest

	// Match Action: ToolName[param1="value1", param2="value2"]
	reactPattern := regexp.MustCompile(`Action:\s*(\w+)\[(.*?)\]`)
	matches := reactPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			toolName := match[1]
			paramsStr := match[2]

			params := parseReActParams(paramsStr)
			requests = append(requests, ToolRequest{
				Name:       toolName,
				Parameters: params,
			})
		}
	}

	return requests
}

// parseReActParams parses ReAct-style parameters
// Example: file_path="/path", limit=10
func parseReActParams(paramsStr string) map[string]interface{} {
	params := make(map[string]interface{})

	// Split by comma
	parts := strings.Split(paramsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Match key="value" or key=value
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.Trim(strings.TrimSpace(kv[1]), "\"'")
				params[key] = value
			}
		}
	}

	return params
}

// GetCompatibilitySystemPrompt returns a system prompt that teaches models
// how to request tools without native function calling support
func GetCompatibilitySystemPrompt() string {
	return `You are an AI coding assistant with access to tools for file operations and command execution.

IMPORTANT: This system doesn't support native function calling. Instead, use XML tags to request tools.

Available Tools:
1. Bash - Execute shell commands
2. Read - Read files
3. Write - Create/overwrite files
4. Edit - Replace text in files
5. Glob - Find files by pattern
6. Grep - Search file contents
7. TodoWrite - Manage task lists

Tool Request Format:
<tool_use name="ToolName">
<param1>value1</param1>
<param2>value2</param2>
</tool_use>

Examples:

To read a file:
<tool_use name="Read">
<file_path>/home/user/README.md</file_path>
</tool_use>

To find files:
<tool_use name="Glob">
<pattern>*.go</pattern>
</tool_use>

To execute a command:
<tool_use name="Bash">
<command>ls -la</command>
</tool_use>

To search for content:
<tool_use name="Grep">
<pattern>TODO</pattern>
<output_mode>files_with_matches</output_mode>
</tool_use>

To write a file:
<tool_use name="Write">
<file_path>/home/user/test.txt</file_path>
<content>Hello World!</content>
</tool_use>

To edit a file:
<tool_use name="Edit">
<file_path>/home/user/file.txt</file_path>
<old_string>old text</old_string>
<new_string>new text</new_string>
</tool_use>

Workflow:
1. When you need to use a tool, output the tool request using the XML format above
2. The system will execute the tool and provide results
3. Use the results to answer the user's question

Always use tools when you need to:
- Read files (don't guess contents)
- List/find files (don't make assumptions)
- Execute commands (don't simulate)
- Search code (use actual search)

Think step by step and use the appropriate tools to complete tasks accurately.`
}
