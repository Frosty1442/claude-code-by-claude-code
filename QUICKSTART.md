# Quick Start Guide

Get up and running with Claude Code Clone in 5 minutes!

## Prerequisites

You need one of:
- OpenAI API key
- Azure OpenAI access
- LocalAI running locally
- Ollama with OpenAI compatibility
- Any OpenAI-compatible API endpoint

## Installation

### Option 1: Download Pre-built Binary (Coming Soon)

```bash
# Download from releases page
# Extract and run
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/claude-code-clone/claude-code-clone.git
cd claude-code-clone

# Build
make build

# Binary will be at: build/claude-code-clone
```

## Configuration

Set your API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

That's it! For OpenAI, this is all you need. For other providers:

```bash
# LocalAI
export OPENAI_API_KEY="dummy"
export OPENAI_BASE_URL="http://localhost:8080/v1"
export MODEL="your-model"

# Ollama
export OPENAI_API_KEY="dummy"
export OPENAI_BASE_URL="http://localhost:11434/v1"
export MODEL="llama2"

# Azure OpenAI
export OPENAI_API_KEY="your-azure-key"
export OPENAI_BASE_URL="https://your-resource.openai.azure.com/..."
export MODEL="gpt-4"
```

## First Run

```bash
./build/claude-code-clone
```

You'll see:
```
Claude Code Clone - Interactive AI Assistant
Type your message and press Enter. Type 'exit' to quit, 'clear' to clear history.
------------------------------------------------------------

>
```

## Example Usage

### Read a File

```
> Read the README.md file
```

The assistant will use the Read tool to show you the file contents.

### Search for Files

```
> Find all Go files in this project
```

The assistant will use the Glob tool to find matching files.

### Search File Contents

```
> Search for "TODO" comments in the code
```

The assistant will use the Grep tool to search.

### Create a File

```
> Create a file called hello.txt with "Hello, World!"
```

The assistant will use the Write tool.

### Edit a File

```
> In hello.txt, replace "World" with "Claude Code Clone"
```

The assistant will use the Edit tool.

### Run a Command

```
> Run 'go version'
```

The assistant will use the Bash tool.

### Multi-step Task

```
> Help me create a new Go module called 'utils' with a function to reverse a string. Create the file, write the code, and test it.
```

The assistant will:
1. Create a todo list to track the task
2. Create the necessary files
3. Write the code
4. Run tests
5. Report back

## Tips

1. **Be specific**: The more specific your request, the better the result
2. **Use natural language**: Just describe what you want
3. **Multiple tasks**: The assistant can handle complex multi-step tasks
4. **File paths**: Use relative paths (from current directory) or absolute paths
5. **Clear history**: Type `clear` to start fresh
6. **Exit**: Type `exit` or `quit` to close

## Common Workflows

### Code Review

```
> Review all Go files in the internal/ directory and suggest improvements
```

### Refactoring

```
> Refactor the main.go file to use better error handling
```

### Documentation

```
> Read all .go files in internal/tools/ and create API documentation
```

### Testing

```
> Create unit tests for all functions in internal/tools/bash.go
```

## Troubleshooting

### "OPENAI_API_KEY environment variable is required"

Solution: Set your API key
```bash
export OPENAI_API_KEY="your-key"
```

### "API error (status 401)"

Solution: Check your API key is valid

### "API error (status 404)"

Solution: Check your BASE_URL and MODEL are correct

### Binary not found

Solution: Build it first
```bash
make build
```

## Next Steps

- Read the full [README.md](README.md) for all features
- Check [ARCHITECTURE.md](ARCHITECTURE.md) to understand the design
- Read [CONTRIBUTING.md](CONTRIBUTING.md) to contribute
- Explore advanced features like task management and parallel execution

## Getting Help

- Type `help` in the REPL
- Check the README for detailed documentation
- Open an issue on GitHub

## Safety Tips

- Always review AI suggestions before accepting
- Use in version-controlled projects (git)
- Don't commit sensitive API keys
- Be careful with destructive operations

Happy coding! 🚀
