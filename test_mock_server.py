#!/usr/bin/env python3
"""
Mock OpenAI-compatible API server for testing
"""
from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import sys

class MockAPIHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/v1/chat/completions':
            content_length = int(self.headers['Content-Length'])
            body = self.rfile.read(content_length)
            request_data = json.loads(body)

            # Log the request
            print(f"\n=== Received Request ===", file=sys.stderr)
            print(f"Messages: {len(request_data.get('messages', []))}", file=sys.stderr)
            print(f"Tools: {len(request_data.get('tools', []))}", file=sys.stderr)

            # Check if last message asks to use a tool
            messages = request_data.get('messages', [])
            if messages:
                last_msg = messages[-1]
                print(f"Last message role: {last_msg.get('role')}", file=sys.stderr)

            # Create a response that uses the Read tool
            response = {
                "id": "test-123",
                "object": "chat.completion",
                "created": 1234567890,
                "model": "test-model",
                "choices": [{
                    "index": 0,
                    "message": {
                        "role": "assistant",
                        "content": "I'll help you read that file.",
                        "tool_calls": [{
                            "id": "call_123",
                            "type": "function",
                            "function": {
                                "name": "Read",
                                "arguments": json.dumps({
                                    "file_path": "/home/user/claude-code-by-claude-code/README.md"
                                })
                            }
                        }]
                    },
                    "finish_reason": "tool_calls"
                }],
                "usage": {
                    "prompt_tokens": 10,
                    "completion_tokens": 20,
                    "total_tokens": 30
                }
            }

            # If we received tool results, respond with final answer
            if any('tool_result' in str(msg.get('content', '')) for msg in messages):
                response['choices'][0]['message'] = {
                    "role": "assistant",
                    "content": "I've successfully read the README.md file and here's what I found. The file contains documentation about the Claude Code Clone project."
                }
                response['choices'][0]['finish_reason'] = "stop"

            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())
            print(f"Sent response with finish_reason: {response['choices'][0]['finish_reason']}", file=sys.stderr)
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        # Suppress default logging
        pass

if __name__ == '__main__':
    port = 8080
    print(f"Starting mock OpenAI API server on port {port}...", file=sys.stderr)
    server = HTTPServer(('localhost', port), MockAPIHandler)
    print(f"Mock server ready at http://localhost:{port}/v1", file=sys.stderr)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down server", file=sys.stderr)
        server.shutdown()
