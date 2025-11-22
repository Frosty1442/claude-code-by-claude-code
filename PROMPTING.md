# System Prompting Strategy

This document explains the comprehensive system prompts that make this Claude Code clone effective.

## Why Good Prompting Matters

The quality of system prompts is **critical** to an AI coding assistant's effectiveness. Good prompts:

1. **Guide tool usage** - Tell the model when and how to use each tool
2. **Establish best practices** - Enforce safe coding patterns (read before edit, secret detection, etc.)
3. **Set tone and style** - Define response format, professionalism, objectivity
4. **Handle edge cases** - Explain error recovery, retry logic, common pitfalls
5. **Optimize performance** - Teach parallel execution, context management

## Our Prompting Approach

### Comprehensive System Prompts

We have **two comprehensive system prompts** totaling 566 lines:

1. **Standard Mode** (`internal/config/system_prompt.go`) - 249 lines
   - For models with native function calling support
   - Detailed instructions for all 11 tools
   - Best practices for code editing, git operations, security
   - Workflow patterns for common tasks
   - Error handling guidance

2. **Compatibility Mode** (`internal/conversation/compatibility.go`) - 317 lines
   - For models WITHOUT function calling (TinyLlama, smaller Mistral, etc.)
   - Teaches XML-based tool requests
   - Complete examples for each tool
   - Same best practices adapted for XML format

### Key Sections in Standard Mode Prompt

#### 1. Core Capabilities
Lists all 11 tools with brief descriptions:
- Bash, BashOutput, KillShell
- Read, Write, Edit
- Glob, Grep
- TodoWrite, Git, WebFetch

#### 2. Critical Rules

**File Operations:**
- ALWAYS read before edit (most important rule!)
- Use exact string matching from Read output
- Preserve indentation exactly
- Never include line numbers in old_string

**Command Execution:**
- Quote paths with spaces
- Parallel vs sequential execution
- Background process management

**Git Operations:**
- NEVER force push to main/master
- NEVER commit secrets (auto-detected)
- Conventional commit messages
- Automatic retry with exponential backoff

**Search and Discovery:**
- When to use Glob vs Grep
- Output modes and context lines
- Efficient search patterns

**Task Management:**
- When to create todos (3+ steps, multiple features)
- Status tracking (pending → in_progress → completed)
- Only ONE in_progress task at a time

#### 3. Response Style

**Tone and Communication:**
- Concise and technical
- Facts over validation
- Objective information
- Professional tone
- Markdown formatting
- File paths with line numbers: `file.go:123`

**Code Quality:**
- Clean, idiomatic code
- Language best practices
- Meaningful names
- Error handling
- Edge cases

**Security Awareness:**
- Never commit sensitive data
- Cautious with destructive operations
- Validate user input
- Avoid injection vulnerabilities
- Use environment variables

#### 4. Workflow Patterns

Complete workflows for:
- **Code Changes**: Read → Grep → Read → Edit → Verify
- **Debugging**: Read → Search → Analyze → Fix → Implement
- **New Features**: Understand → Check structure → Read context → TodoWrite → Implement → Test
- **Git Commits**: Status → Diff → Stage → Commit → Push (only if requested)

#### 5. Error Handling

**Common Issues:**
- Edit fails (file not read) → Read the file
- Edit fails (string not found) → Verify exact string, add context
- Bash fails (path with spaces) → Quote the path
- Git fails (network) → Auto-retry, wait
- WebFetch fails (invalid URL) → Check format

#### 6. Performance Optimization

- Parallel tool execution when possible
- Minimize file reads
- Use Grep before reading large files
- Efficient Glob patterns
- Cache frequently accessed info
- Consider token usage

#### 7. Final Principles

10 core principles:
1. Accuracy over speed
2. Read before write
3. Verify operations
4. Security first
5. Clear communication
6. Professional tone
7. Handle errors gracefully
8. Think step-by-step
9. Use tools effectively
10. Learn from feedback

### Compatibility Mode Prompt

For models without function calling, we teach XML-based tool requests:

```xml
<tool_use name="Read">
<file_path>/path/to/file</file_path>
</tool_use>
```

The compatibility prompt includes:
- **XML Format Specification** - Exact syntax with examples
- **All 11 Tools** - With parameter descriptions
- **Complete Examples** - For every tool
- **Critical Rules** - Same safety rules as standard mode
- **Workflow Patterns** - Adapted for XML requests

## Comparison with Claude Code and Cline

### Claude Code
- **Official Anthropic implementation**
- Comprehensive prompts with tool instructions
- Professional tone and objectivity
- Security best practices
- Error handling guidance

### Cline
- **Popular VSCode extension**
- Detailed prompts for code editing
- VSCode-specific integrations
- Task management with todos
- Git workflow guidance

### Our Implementation
- **Complete clone** of prompting best practices
- 566 lines of detailed instructions
- Covers all 11 tools comprehensively
- Both function calling and XML modes
- Security-first approach
- Production-ready error handling

## Testing the Prompts

### Standard Mode Test
```bash
./build/claude-code-clone
> Read the README.md file
```

The model will:
1. Use the Read tool correctly
2. Respond professionally
3. Format output clearly

### Compatibility Mode Test
```bash
./build/claude-code-clone --compatibility
> Read the README.md file
```

The model will:
1. Output XML tool request
2. System executes and returns result
3. Model responds with the content

## Why This Matters

Without comprehensive prompts:
- ❌ Models don't know when to read before editing
- ❌ Edits fail due to inexact string matching
- ❌ Secrets get committed to git
- ❌ File paths with spaces cause errors
- ❌ Models don't use parallel execution
- ❌ Response style is inconsistent
- ❌ Error recovery is poor

With comprehensive prompts:
- ✅ Models always read before editing
- ✅ Exact string matching from Read output
- ✅ Automatic secret detection prevents leaks
- ✅ Proper path quoting
- ✅ Parallel execution when possible
- ✅ Professional, concise responses
- ✅ Graceful error handling

## Future Enhancements

Possible improvements to prompts:
1. **Language-specific guidance** - Python vs Go vs JavaScript best practices
2. **Testing patterns** - How to write tests, run test suites
3. **Code review guidelines** - Security, performance, maintainability checks
4. **Refactoring patterns** - Safe ways to restructure code
5. **Documentation standards** - When and how to add comments/docs
6. **Performance profiling** - When to use profilers, how to interpret results

## Customization

Users can customize system prompts by:

1. **Editing system_prompt.go** - Modify the GetSystemPrompt() function
2. **Environment variable** - Set SYSTEM_PROMPT env var (future enhancement)
3. **YAML config** - Add system_prompt field to .claude/config.yaml (future enhancement)

## References

- [Claude Code Documentation](https://docs.claude.com/docs/claude-code)
- [Cline GitHub](https://github.com/cline/cline)
- [Anthropic Prompt Engineering Guide](https://docs.anthropic.com/claude/docs/prompt-engineering)
