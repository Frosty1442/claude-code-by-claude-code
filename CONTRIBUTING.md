# Contributing to Claude Code Clone

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/claude-code-clone.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `make test`
6. Commit your changes: `git commit -am 'Add some feature'`
7. Push to the branch: `git push origin feature/your-feature-name`
8. Create a Pull Request

## Development Setup

### Prerequisites
- Go 1.21 or later
- Make (optional but recommended)

### Building
```bash
make build
```

### Testing
```bash
make test
```

### Code Formatting
```bash
make fmt
```

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Write meaningful commit messages
- Add tests for new features
- Update documentation as needed

## Project Structure

```
claude-code-clone/
├── cmd/              # Main applications
├── internal/         # Private application code
│   ├── api/         # API clients
│   ├── tools/       # Tool implementations
│   ├── conversation/ # Conversation management
│   ├── repl/        # REPL interface
│   ├── executor/    # Tool execution
│   └── config/      # Configuration
├── pkg/             # Public libraries
└── vendor/          # Vendored dependencies
```

## Adding New Tools

To add a new tool:

1. Create a new file in `internal/tools/`
2. Implement the `Tool` interface:
   ```go
   type Tool interface {
       Name() string
       Description() string
       Schema() schema.ToolDefinition
       Execute(params map[string]interface{}) (*ToolResult, error)
       Validate(params map[string]interface{}) error
   }
   ```
3. Add the tool schema to `pkg/schema/tools.go`
4. Register the tool in `internal/tools/tool.go` in `NewDefaultRegistry()`
5. Add tests in `internal/tools/tool_test.go`

## Testing Guidelines

- Write unit tests for all new code
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example:
```go
func TestMyFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Documentation

- Update README.md for user-facing changes
- Add godoc comments for exported functions
- Update ARCHITECTURE.md for design changes
- Include examples in documentation

## Pull Request Process

1. Ensure all tests pass
2. Update documentation
3. Add a clear PR description
4. Link related issues
5. Wait for review
6. Address feedback
7. Maintainer will merge when approved

## Reporting Issues

When reporting issues, please include:

- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment (OS, Go version, etc.)
- Logs or error messages
- Minimal reproducible example

## Feature Requests

Feature requests are welcome! Please:

- Check if it already exists in issues
- Describe the use case
- Explain why it would be useful
- Provide examples if possible

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Assume good intentions

## Questions?

Feel free to:
- Open an issue for questions
- Start a discussion on GitHub
- Reach out to maintainers

Thank you for contributing!
