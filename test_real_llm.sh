#!/bin/bash
export OPENAI_API_KEY="not-needed"
export OPENAI_BASE_URL="http://127.0.0.1:8080/v1"
export MODEL="/tmp/models/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf"

cd /home/user/claude-code-by-claude-code

echo "Testing with REAL TinyLlama LLM..."
echo ""
echo "Query: List all Go files in this project"
echo ""

echo "List all Go files in this project" | timeout 60 ./build/claude-code-clone 2>&1
