#!/bin/bash
# Interactive test with timeout

export OPENAI_API_KEY="test-key"
export OPENAI_BASE_URL="http://localhost:8080/v1"
export MODEL="test-model"

echo "=== Claude Code Clone Test ==="
echo ""

# Test 1: Simple message
echo "Test 1: Read README.md"
echo "Read the README.md file" | timeout 5 ./build/claude-code-clone 2>&1

echo ""
echo "===================="
