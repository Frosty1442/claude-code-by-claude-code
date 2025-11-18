# Makefile for Claude Code Clone

.PHONY: all build build-offline install clean test vendor help

# Variables
BINARY_NAME=claude-code-clone
MAIN_PATH=./cmd/claude-code
BUILD_DIR=./build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Default target
all: build

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build with vendored dependencies (offline)
build-offline: vendor
	@echo "Building $(BINARY_NAME) with vendored dependencies..."
	@mkdir -p $(BUILD_DIR)
	go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Offline build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Vendor dependencies for offline builds
vendor:
	@echo "Vendoring dependencies..."
	go mod vendor
	@echo "Dependencies vendored to ./vendor"

# Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(MAIN_PATH)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf vendor
	@go clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Cross-compile for multiple platforms
build-all: vendor
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)

	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

	@echo "All builds complete!"
	@ls -lh $(BUILD_DIR)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies downloaded"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Dependencies tidied"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

# Help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make build-offline  - Build with vendored dependencies (offline)"
	@echo "  make vendor         - Vendor dependencies"
	@echo "  make install        - Install to GOPATH/bin"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make build-all      - Cross-compile for all platforms"
	@echo "  make deps           - Download dependencies"
	@echo "  make tidy           - Tidy dependencies"
	@echo "  make run            - Run the application"
	@echo "  make help           - Show this help message"
