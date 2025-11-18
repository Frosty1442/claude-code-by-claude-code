#!/bin/bash
# Test script for Claude Code Clone

export OPENAI_API_KEY="test-key"
export OPENAI_BASE_URL="http://localhost:8080/v1"
export MODEL="test-model"

echo "Testing Claude Code Clone with mock API..."
echo ""
echo "Configuration:"
echo "  API URL: $OPENAI_BASE_URL"
echo "  Model: $MODEL"
echo ""
echo "Sending test message: 'Read the README.md file'"
echo ""

# Send a test message via stdin
echo "Read the README.md file" | timeout 10 ./build/claude-code-clone 2>&1
