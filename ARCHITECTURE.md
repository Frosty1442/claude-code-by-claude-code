# System Architecture

## Directory Structure

```
claude-code-clone/
├── cmd/
│   └── claude-code/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── api/
│   │   ├── client.go              # LLM API client interface
│   │   ├── openai.go              # OpenAI-compatible implementation
│   │   ├── anthropic.go           # Native Anthropic implementation
│   │   └── streaming.go           # Streaming response handling
│   ├── tools/
│   │   ├── tool.go                # Tool interface definition
│   │   ├── bash.go                # Bash tool
│   │   ├── read.go                # Read tool
│   │   ├── write.go               # Write tool
│   │   ├── edit.go                # Edit tool
│   │   ├── glob.go                # Glob tool
│   │   ├── grep.go                # Grep tool
│   │   ├── webfetch.go            # WebFetch tool
│   │   ├── websearch.go           # WebSearch tool
│   │   ├── todo.go                # TodoWrite tool
│   │   ├── notebook.go            # NotebookEdit tool
│   │   ├── task.go                # Task/Agent tool
│   │   ├── skill.go               # Skill tool
│   │   ├── slashcmd.go            # SlashCommand tool
│   │   ├── bashoutput.go          # BashOutput tool
│   │   └── killshell.go           # KillShell tool
│   ├── agent/
│   │   ├── agent.go               # Agent interface
│   │   ├── manager.go             # Agent lifecycle management
│   │   ├── explorer.go            # Explore agent
│   │   ├── planner.go             # Plan agent
│   │   └── general.go             # General-purpose agent
│   ├── conversation/
│   │   ├── manager.go             # Conversation state management
│   │   ├── message.go             # Message types and formatting
│   │   ├── context.go             # Context tracking
│   │   └── history.go             # History persistence
│   ├── config/
│   │   ├── config.go              # Configuration loading
│   │   ├── hooks.go               # Hook management
│   │   ├── skills.go              # Skill loading
│   │   └── commands.go            # Slash command loading
│   ├── git/
│   │   ├── git.go                 # Git operations
│   │   ├── commit.go              # Smart commit creation
│   │   ├── pr.go                  # Pull request handling
│   │   └── safety.go              # Safety checks
│   ├── repl/
│   │   ├── repl.go                # Interactive REPL
│   │   ├── input.go               # Input handling
│   │   └── display.go             # Output rendering
│   ├── executor/
│   │   ├── executor.go            # Tool execution orchestration
│   │   ├── parallel.go            # Parallel execution
│   │   └── sandbox.go             # Sandbox mode
│   ├── fileops/
│   │   ├── reader.go              # File reading (text, image, PDF)
│   │   ├── writer.go              # File writing
│   │   ├── editor.go              # File editing
│   │   ├── glob.go                # Pattern matching
│   │   └── grep.go                # Content searching
│   └── util/
│       ├── markdown.go            # Markdown rendering
│       ├── json.go                # JSON utilities
│       ├── retry.go               # Retry logic
│       └── platform.go            # Platform detection
├── pkg/
│   └── schema/
│       ├── tools.go               # Tool JSON schemas
│       └── messages.go            # Message schemas
├── vendor/                        # Vendored dependencies
├── .claude/
│   ├── commands/                  # Slash commands
│   └── skills/                    # Skills
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── ARCHITECTURE.md
```

## Component Design

### 1. API Client Layer (`internal/api`)

**Purpose**: Abstract LLM API interactions

**Components**:
- `Client` interface: Common interface for all LLM providers
- `OpenAIClient`: OpenAI-compatible endpoint implementation
- `AnthropicClient`: Native Anthropic API implementation
- `StreamHandler`: Server-sent events (SSE) parsing

**Key Features**:
- Streaming response support
- Automatic retry with exponential backoff
- Token counting and budget tracking
- Context window management
- Model selection (sonnet, opus, haiku equivalents)

### 2. Tool System (`internal/tools`)

**Purpose**: Implement all Claude Code tools

**Interface**:
```go
type Tool interface {
    Name() string
    Description() string
    Schema() map[string]interface{}
    Execute(params map[string]interface{}) (ToolResult, error)
    Validate(params map[string]interface{}) error
}

type ToolResult struct {
    Output string
    Error  error
    Type   string // "success", "error", "system"
}
```

**Tool Categories**:
1. **File Tools**: Read, Write, Edit, Glob, Grep, NotebookEdit
2. **Shell Tools**: Bash, BashOutput, KillShell
3. **Web Tools**: WebFetch, WebSearch
4. **Meta Tools**: Task, Skill, SlashCommand, TodoWrite

**Design Principles**:
- Each tool is self-contained
- Validation before execution
- Comprehensive error handling
- Sandbox mode support

### 3. Agent System (`internal/agent`)

**Purpose**: Multi-agent task execution

**Components**:
- `Agent`: Base agent interface
- `AgentManager`: Lifecycle and coordination
- Specialized agents: Explorer, Planner, General

**Features**:
- Parallel agent execution
- Context isolation
- Result aggregation
- Timeout management

### 4. Conversation Manager (`internal/conversation`)

**Purpose**: Manage conversation state and flow

**Responsibilities**:
- Message history tracking
- Context window management
- Tool call parsing and formatting
- Token budget tracking
- History persistence (optional)

**Message Flow**:
```
User Input → REPL → Conversation Manager → API Client → LLM
                ↑                                         ↓
                └──────── Tool Executor ←─────── Tool Calls
```

### 5. Executor (`internal/executor`)

**Purpose**: Orchestrate tool execution

**Features**:
- Parallel execution detection
- Dependency resolution
- Sandbox mode enforcement
- Result aggregation
- Error recovery

**Execution Modes**:
- Sequential: Tools with dependencies
- Parallel: Independent tools
- Background: Long-running commands

### 6. REPL (`internal/repl`)

**Purpose**: Interactive user interface

**Features**:
- Multi-line input support
- Command history
- Streaming output display
- Markdown rendering
- Syntax highlighting
- Progress indicators

### 7. Configuration (`internal/config`)

**Purpose**: Load and manage configuration

**Configuration Sources**:
1. Environment variables
2. Config file (.claude/config.yaml)
3. CLI flags
4. Runtime settings

**Managed Elements**:
- API keys and endpoints
- Model preferences
- Hooks (session-start, prompt-submit)
- Slash commands (.claude/commands/*.md)
- Skills (.claude/skills/*.md)
- Sandbox settings

### 8. Git Integration (`internal/git`)

**Purpose**: Safe and smart git operations

**Features**:
- Status, diff, log parsing
- Smart commit message generation
- Pre-commit hook handling
- Branch management
- PR creation
- Safety checks (force-push protection, authorship validation)

### 9. File Operations (`internal/fileops`)

**Purpose**: High-performance file handling

**Capabilities**:
- Fast glob matching (doublestar)
- Ripgrep-compatible content search
- Multimodal file reading (images via base64, PDFs via extraction)
- Safe file editing (atomic writes, backups)
- Jupyter notebook parsing

## Data Flow

### Complete Request Flow

```
1. User Input
   ↓
2. REPL (parse input, check for slash commands)
   ↓
3. Hooks (user-prompt-submit-hook)
   ↓
4. Conversation Manager (add to history, format for API)
   ↓
5. API Client (send to LLM with tools schema)
   ↓
6. Stream Response
   ↓
7. Parse Tool Calls
   ↓
8. Executor (execute tools in parallel/sequential)
   ↓
9. Tool Results
   ↓
10. Conversation Manager (add results to history)
    ↓
11. API Client (continue conversation with results)
    ↓
12. Display Final Response
```

## Key Design Decisions

### 1. Tool Execution Strategy
- **Parallel by default**: Detect independent tools and run concurrently
- **Dependency detection**: Analyze tool parameters for references to other tool outputs
- **Timeout handling**: Per-tool timeouts with graceful termination

### 2. Sandbox Mode
- **Configurable**: Enable/disable via config
- **Whitelisted operations**: Only allow safe file/command operations
- **Path validation**: Restrict to working directory

### 3. Error Handling
- **Graceful degradation**: Continue on non-critical errors
- **Retry logic**: Exponential backoff for network operations
- **User feedback**: Clear, actionable error messages

### 4. State Management
- **Stateless tools**: Each tool execution is independent
- **Persistent conversation**: History maintained in memory + optional disk
- **Background shells**: State tracked in shell manager

### 5. Extensibility
- **Plugin architecture**: Easy to add new tools
- **Hook system**: Customizable behavior at key points
- **Configuration-driven**: Behavior controlled by config files

## Performance Considerations

### 1. Concurrency
- Goroutines for parallel tool execution
- Channels for result aggregation
- Context for cancellation

### 2. Caching
- WebFetch: 15-minute cache
- File reads: No caching (always fresh)
- Glob/Grep: No caching (fast enough)

### 3. Resource Limits
- Max concurrent tools: 10
- Max background shells: 5
- Max token budget: Configurable
- Max file size: 10MB (configurable)

## Security Considerations

### 1. Command Injection Prevention
- Parameterized execution (no shell interpolation)
- Input validation
- Whitelist of allowed commands in sandbox mode

### 2. Path Traversal Prevention
- Absolute path resolution
- Working directory enforcement
- Symlink validation

### 3. Secret Detection
- Scan for common secret patterns before commits
- Warn on sensitive file commits (.env, credentials.json)

### 4. API Key Management
- Environment variables preferred
- Config file with restricted permissions
- Never log API keys

## Testing Strategy

### 1. Unit Tests
- Each tool independently
- Mock API responses
- Edge case coverage

### 2. Integration Tests
- End-to-end conversation flows
- Tool chaining
- Error scenarios

### 3. Performance Tests
- Parallel execution benchmarks
- Large file handling
- Memory usage profiling
