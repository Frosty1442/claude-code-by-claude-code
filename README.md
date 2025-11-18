# Claude Code Clone

A full-featured AI coding assistant inspired by Claude Code, implementing similar functionality with support for OpenAI-compatible API endpoints.

## Features

### Core Capabilities
- **File Operations**: Read, write, and edit files with precision
- **Pattern Matching**: Fast glob-based file finding
- **Content Search**: Powerful grep with regex support
- **Command Execution**: Run bash commands safely
- **Task Management**: Built-in todo list tracking
- **Conversation Management**: Multi-turn conversations with context
- **Streaming Responses**: Real-time API response streaming

### Tool System
- **Bash**: Execute shell commands with timeout support
- **Read**: Read any file (text, images via base64)
- **Write**: Create or overwrite files
- **Edit**: Exact string replacement in files
- **Glob**: Find files by pattern (supports `**` for recursive)
- **Grep**: Search file contents with regex
  - Multiple output modes: content, files_with_matches, count
  - Context lines support (-A, -B, -C)
  - File type filtering
- **TodoWrite**: Manage structured task lists

### OpenAI API Compatible
- Works with any OpenAI-compatible endpoint
- Supports GPT-4, GPT-3.5, and similar models
- Compatible with:
  - OpenAI
  - Azure OpenAI
  - LocalAI
  - Ollama (with OpenAI compatibility layer)
  - Any other OpenAI-compatible API

## Installation

### Prerequisites
- Go 1.21 or later (for building)
- API key for OpenAI-compatible service

### Build from Source

```bash
# Clone the repository
git clone https://github.com/claude-code-clone/claude-code-clone.git
cd claude-code-clone

# Build (requires internet for dependencies)
make build

# OR build offline (after first time)
make vendor
make build-offline
```

### Cross-Platform Builds

```bash
# Build for all platforms
make build-all

# Binaries will be in ./build/
# - claude-code-clone-linux-amd64
# - claude-code-clone-linux-arm64
# - claude-code-clone-darwin-amd64
# - claude-code-clone-darwin-arm64
# - claude-code-clone-windows-amd64.exe
```

### Install to System

```bash
make install
# Binary will be installed to $GOPATH/bin/claude-code-clone
```

## Configuration

Set the following environment variables:

```bash
# Required
export OPENAI_API_KEY="your-api-key-here"

# Optional
export OPENAI_BASE_URL="https://api.openai.com/v1"  # Default
export MODEL="gpt-4"                                 # Default
```

### Using with Other Providers

#### Azure OpenAI
```bash
export OPENAI_API_KEY="your-azure-key"
export OPENAI_BASE_URL="https://your-resource.openai.azure.com/openai/deployments/your-deployment"
export MODEL="gpt-4"
```

#### LocalAI
```bash
export OPENAI_API_KEY="dummy"  # LocalAI doesn't require a real key
export OPENAI_BASE_URL="http://localhost:8080/v1"
export MODEL="your-model-name"
```

#### Ollama (with OpenAI compatibility)
```bash
# Start Ollama with OpenAI compatibility
# ollama serve

export OPENAI_API_KEY="dummy"
export OPENAI_BASE_URL="http://localhost:11434/v1"
export MODEL="llama2"
```

## Usage

### Interactive Mode (REPL)

```bash
./build/claude-code-clone
```

Then interact with the assistant:

```
> Read the README.md file

> Search for all Go files in the project

> Find TODO comments in the code

> Create a new file called test.txt with "Hello World"

> Edit main.go to add a new function
```

### Commands

- `help` - Show help message
- `clear` - Clear conversation history
- `exit` or `quit` - Exit the program

### Example Interactions

#### File Operations
```
> Read the package.json file
> Create a new file src/utils.js with a helper function
> Edit config.yaml to change the port to 8080
```

#### Code Search
```
> Find all TypeScript files
> Search for "TODO" comments in the codebase
> Find files containing "API_KEY"
```

#### Task Management
```
> Help me implement a new authentication system. Create a todo list for this task.
```

The assistant will create and track tasks as it works.

## Architecture

```
claude-code-clone/
├── cmd/
│   └── claude-code/          # Main entry point
├── internal/
│   ├── api/                  # API client (OpenAI-compatible)
│   ├── tools/                # Tool implementations
│   ├── conversation/         # Conversation management
│   ├── repl/                 # Interactive REPL
│   ├── executor/             # Tool execution orchestration
│   └── config/               # Configuration management
├── pkg/
│   └── schema/               # Shared schemas and types
├── vendor/                   # Vendored dependencies (offline builds)
└── build/                    # Build output
```

### Key Components

1. **API Client**: Abstraction over OpenAI-compatible APIs
   - Streaming support
   - Automatic retry with exponential backoff
   - Token tracking

2. **Tool System**: Pluggable tool architecture
   - Each tool is self-contained
   - Validation and error handling
   - Parallel execution support

3. **Conversation Manager**: Handles conversation flow
   - Message history management
   - Tool call execution loop
   - Context window management

4. **REPL**: Interactive command-line interface
   - Multi-line input support
   - Command history
   - Streaming output

## Development

### Running Tests

```bash
make test

# With coverage
make test-coverage
```

### Code Formatting

```bash
make fmt
```

### Linting

```bash
make lint
```

### Building

```bash
# Standard build
make build

# Offline build (with vendored dependencies)
make vendor
make build-offline

# Cross-platform build
make build-all
```

## Offline Usage

This project is designed to work completely offline after initial setup:

1. First time (requires internet):
```bash
make vendor
```

2. Subsequent builds (no internet required):
```bash
make build-offline
```

All dependencies are vendored in the `vendor/` directory.

## Comparison with Similar Tools

| Feature | Claude Code Clone | Cline | Goose | Continue |
|---------|------------------|-------|-------|----------|
| OpenAI Compatible | ✅ | ❌ | ✅ | ✅ |
| CLI Interface | ✅ | ❌ | ✅ | ❌ |
| VSCode Extension | ❌ | ✅ | ❌ | ✅ |
| File Operations | ✅ | ✅ | ✅ | ✅ |
| Task Management | ✅ | ✅ | ❌ | ❌ |
| Offline Build | ✅ | ❌ | ❌ | ❌ |
| Single Binary | ✅ | ❌ | ✅ | ❌ |

## Roadmap

Future enhancements (not yet implemented):
- [ ] Web search and web fetch capabilities
- [ ] Agent/task management system
- [ ] Git integration
- [ ] Jupyter notebook support
- [ ] Slash commands and skills system
- [ ] Hooks system (session-start, prompt-submit)
- [ ] MCP (Model Context Protocol) support
- [ ] Streaming token-by-token output

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

Inspired by:
- [Claude Code](https://claude.ai/code) - Anthropic's official CLI
- [Cline](https://github.com/cline/cline) - VSCode extension
- [Goose](https://github.com/block/goose) - AI coding assistant
- [Continue](https://continue.dev/) - AI code assistant

## Support

For issues and questions:
- GitHub Issues: https://github.com/claude-code-clone/claude-code-clone/issues

## Security

This tool executes commands and modifies files based on AI suggestions. Always:
- Review changes before confirming
- Use in a version-controlled environment
- Don't expose API keys in code
- Be cautious with sensitive files

## API Costs

This tool makes API calls to OpenAI-compatible services. Be aware of:
- Token usage is displayed after each response
- Costs vary by provider and model
- Consider using local models (LocalAI, Ollama) for cost savings
