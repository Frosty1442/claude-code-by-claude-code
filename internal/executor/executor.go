package executor

import (
	"sync"

	"github.com/claude-code-clone/claude-code-clone/internal/tools"
	"github.com/claude-code-clone/claude-code-clone/pkg/schema"
)

// Executor orchestrates tool execution
type Executor struct {
	registry *tools.Registry
}

// NewExecutor creates a new executor
func NewExecutor(registry *tools.Registry) *Executor {
	return &Executor{
		registry: registry,
	}
}

// ExecuteParallel executes multiple tools in parallel
func (e *Executor) ExecuteParallel(toolCalls []schema.ContentBlock) []ToolExecutionResult {
	results := make([]ToolExecutionResult, len(toolCalls))
	var wg sync.WaitGroup

	for i, toolCall := range toolCalls {
		wg.Add(1)
		go func(idx int, tc schema.ContentBlock) {
			defer wg.Done()
			result := e.executeSingle(tc)
			results[idx] = result
		}(i, toolCall)
	}

	wg.Wait()
	return results
}

// ExecuteSequential executes tools one by one
func (e *Executor) ExecuteSequential(toolCalls []schema.ContentBlock) []ToolExecutionResult {
	results := make([]ToolExecutionResult, len(toolCalls))

	for i, toolCall := range toolCalls {
		results[i] = e.executeSingle(toolCall)
	}

	return results
}

// executeSingle executes a single tool
func (e *Executor) executeSingle(toolCall schema.ContentBlock) ToolExecutionResult {
	result, err := e.registry.Execute(toolCall.Name, toolCall.Input)

	return ToolExecutionResult{
		ToolID:  toolCall.ID,
		Name:    toolCall.Name,
		Output:  result.Output,
		Error:   err,
		IsError: result.Type == "error",
	}
}

// ToolExecutionResult represents the result of a tool execution
type ToolExecutionResult struct {
	ToolID  string
	Name    string
	Output  string
	Error   error
	IsError bool
}

// ToContentBlock converts to a content block
func (r *ToolExecutionResult) ToContentBlock() schema.ContentBlock {
	return schema.NewToolResultBlock(r.ToolID, r.Output, r.IsError)
}
