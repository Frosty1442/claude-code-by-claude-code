# Claude Code Clone - Complete Feature List

## 1. Core Tool System

### File Operations
- **Read**: Read any file (text, images, PDFs, Jupyter notebooks)
  - Support for line offset and limit
  - Multimodal support (images, PDFs)
  - Jupyter notebook cell parsing
- **Write**: Create or overwrite files
- **Edit**: Exact string replacement in files
  - Single replacement mode
  - Replace-all mode for renaming
- **Glob**: Pattern-based file finding
  - Supports standard glob patterns (*.js, **/*.tsx)
  - Sorted by modification time
- **Grep**: Content search with ripgrep
  - Full regex support
  - Multiple output modes (content, files_with_matches, count)
  - Context lines (-A, -B, -C)
  - Case-insensitive search
  - Multiline mode
  - Type filtering (js, py, rust, etc.)
- **NotebookEdit**: Edit Jupyter notebook cells
  - Replace, insert, delete modes
  - Cell type management (code/markdown)

### Shell Operations
- **Bash**: Execute shell commands
  - Persistent shell sessions
  - Background execution support
  - Configurable timeouts (up to 10 minutes)
  - Output capture and truncation
- **BashOutput**: Read output from background shells
  - Incremental output reading
  - Regex filtering
- **KillShell**: Terminate background shells

### Web Operations
- **WebFetch**: Fetch and process web content
  - HTML to markdown conversion
  - AI-powered content extraction
  - 15-minute cache
  - Redirect handling
- **WebSearch**: Search the web
  - Domain filtering (allow/block)
  - Current event awareness

### Task Management
- **TodoWrite**: Task tracking and management
  - Three states: pending, in_progress, completed
  - Active form for in-progress display
  - Real-time updates

## 2. Agent System

### Agent Types
- **general-purpose**: Multi-step task automation
- **Explore**: Fast codebase exploration
  - Thoroughness levels: quick, medium, very thorough
- **Plan**: Planning and design tasks
- **statusline-setup**: Configuration management

### Agent Features
- Parallel agent execution
- Stateless operation
- Context-aware agents
- Autonomous task completion
- Sub-agent spawning

## 3. LLM Integration

### API Features
- Support for multiple providers (Claude, OpenAI-compatible)
- Model selection (sonnet, opus, haiku)
- Streaming responses
- Token tracking and budget management
- Rate limiting and retry logic

### Message Handling
- Conversation history management
- System prompts and instructions
- Tool use formatting
- Multimodal message support

## 4. Git Integration

### Core Git Features
- Status, diff, log commands
- Branch management (create, switch, delete)
- Commit creation with smart messages
  - HEREDOC formatting
  - Style inference from history
  - Pre-commit hook handling
- Remote operations (push, pull, fetch)
  - Retry logic with exponential backoff
  - Branch-specific operations

### GitHub Integration
- Pull request creation via gh CLI
- Issue management
- PR comment viewing
- Checks and releases

### Safety Features
- No config modification
- No destructive operations without confirmation
- Hook preservation
- Authorship checking for amends
- Force-push protection for main/master

## 5. Configuration System

### Slash Commands
- Custom commands in .claude/commands/
- Markdown-based command definitions
- Argument support
- Sequential execution

### Skills
- User and system skills
- Scoped execution (project, gitignored)
- Specialized capabilities

### Hooks
- session-start-hook: Initialization tasks
- user-prompt-submit-hook: Pre-submission validation
- Custom hook support

### Settings
- Model preferences
- Token budgets
- Sandbox mode
- Working directory management

## 6. CLI Interface

### REPL Features
- Interactive command loop
- Multi-line input support
- Command history
- Clear and help commands
- Streaming output display

### Display Features
- GitHub-flavored markdown rendering
- Monospace formatting
- Color support
- Progress indicators
- Token usage display

## 7. Advanced Features

### Parallel Execution
- Multiple tool calls in single message
- Independent operation detection
- Sequential dependency handling

### Context Management
- Working directory tracking
- Environment variable access
- Git repository detection
- Platform and OS awareness
- Date/time awareness

### Error Handling
- Graceful failure recovery
- Retry logic for network operations
- User-friendly error messages
- System reminders for best practices

### Security Features
- Sandbox mode for safe execution
- Path validation
- Secret detection in commits
- Command injection prevention
- OWASP top 10 awareness

## 8. File Format Support

- Text files (all encodings)
- Images (PNG, JPG, etc.) - multimodal
- PDFs - page-by-page extraction
- Jupyter notebooks (.ipynb)
- Markdown with CommonMark
- Source code (all languages)

## 9. Development Workflow Features

### Code Quality
- Pre-commit hook integration
- Test execution support
- Build verification
- Type error detection and fixing

### Documentation
- Automatic README generation
- Code explanation
- Comment generation
- API documentation

### Refactoring
- Multi-file renaming
- Code restructuring
- Pattern replacement
- Dependency updates

## 10. User Experience

### Professional Features
- Concise, technical communication
- No unnecessary emojis
- Objective guidance
- Error-first thinking
- Proactive tool usage

### Productivity Features
- Automatic task planning
- Progress tracking
- Completion verification
- Context preservation
- Minimal user interruption

## Implementation Requirements

### Language: Go
- Single binary compilation
- Cross-platform support
- Offline build capability (vendored dependencies)
- Fast execution
- Standard library utilization

### Architecture
- Modular tool system
- Plugin-based extensibility
- Clean separation of concerns
- Testable components
- Configuration-driven behavior

### Build System
- Go modules for dependency management
- Vendor directory for offline builds
- Make or task-based build
- Cross-compilation support
- Version embedding
