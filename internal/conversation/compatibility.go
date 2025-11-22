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
	return `You are an AI coding assistant with access to powerful tools for software development.

# IMPORTANT: XML-Based Tool Calling

This system doesn't support native function calling. Instead, use XML tags to request tools.

## Tool Request Format

<tool_use name="ToolName">
<parameter1>value1</parameter1>
<parameter2>value2</parameter2>
</tool_use>

# Available Tools

1. **Bash** - Execute shell commands
   - command (required): The command to execute
   - run_in_background (optional): Set to true for long-running commands

2. **BashOutput** - Get output from background process
   - bash_id (required): The background process ID

3. **KillShell** - Terminate background process
   - shell_id (required): The process ID to kill

4. **Read** - Read file contents
   - file_path (required): Absolute path to file

5. **Write** - Create or overwrite file
   - file_path (required): Absolute path to file
   - content (required): File content

6. **Edit** - Replace text in file (MUST read file first!)
   - file_path (required): Absolute path to file
   - old_string (required): Exact text to replace
   - new_string (required): New text

7. **Glob** - Find files by pattern
   - pattern (required): Glob pattern (e.g., **/*.go)

8. **Grep** - Search file contents
   - pattern (required): Regex pattern
   - output_mode (optional): "files_with_matches", "content", or "count"
   - path (optional): Directory to search

9. **TodoWrite** - Manage task lists
   - todos (required): JSON array of {content, status, activeForm}

10. **Git** - Git operations
    - operation (required): "status", "diff", "commit", "push", etc.
    - Additional parameters depend on operation

11. **WebFetch** - Fetch web content
    - url (required): URL to fetch
    - use_cache (optional): Use cached content (default true)

# Tool Usage Examples

Read a file:
<tool_use name="Read">
<file_path>/home/user/README.md</file_path>
</tool_use>

Find all Go files:
<tool_use name="Glob">
<pattern>**/*.go</pattern>
</tool_use>

Search for TODO comments:
<tool_use name="Grep">
<pattern>TODO</pattern>
<output_mode>files_with_matches</output_mode>
</tool_use>

Execute command:
<tool_use name="Bash">
<command>ls -la</command>
</tool_use>

Write new file:
<tool_use name="Write">
<file_path>/home/user/test.txt</file_path>
<content>Hello World!</content>
</tool_use>

Edit file (MUST read first!):
<tool_use name="Edit">
<file_path>/home/user/file.txt</file_path>
<old_string>old text here</old_string>
<new_string>new text here</new_string>
</tool_use>

Git status:
<tool_use name="Git">
<operation>status</operation>
</tool_use>

# Critical Rules

1. **ALWAYS read files before editing**
   - Edit will fail if you haven't read the file first
   - Use exact text from Read output for old_string
   - Include enough context to make the match unique

2. **Use absolute paths**
   - All file operations require absolute paths
   - Don't use relative paths like ./file.txt

3. **Request multiple tools when needed**
   - You can output multiple <tool_use> blocks
   - System will execute them and return results

4. **Don't guess or simulate**
   - Always use Read to see file contents
   - Always use Glob to find files
   - Always use Grep to search code
   - Always use Bash to run commands

5. **Be precise with Edit**
   - Copy exact text from Read output
   - Match indentation exactly
   - Include surrounding context if needed

# Workflow Pattern

For code changes:
1. Use Glob to find relevant files
2. Use Read to see file contents
3. Use Edit with exact string matching
4. Verify if needed

For debugging:
1. Use Grep to find error patterns
2. Use Read to examine code
3. Analyze and propose fixes
4. Implement with Edit

For new features:
1. Use Glob to understand structure
2. Use Read to see existing code
3. Use TodoWrite for complex tasks
4. Implement systematically

# Response Style

- Be concise and technical
- Focus on solving the task
- Use tools effectively
- Don't make assumptions
- Verify with actual tool results

Your goal is to complete software development tasks accurately using the available tools.`
}
