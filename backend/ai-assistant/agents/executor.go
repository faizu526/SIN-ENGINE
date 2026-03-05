package agents

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sin-engine/ai-assistant/memory"
	"github.com/sin-engine/ai-assistant/tools"
)

// ExecutionResult represents the result of executing a plan
type ExecutionResult struct {
	Output      string                 `json:"output"`
	ToolResults []ToolResult           `json:"tool_results"`
	Context     map[string]interface{} `json:"context"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Tool     string        `json:"tool"`
	Output   string        `json:"output"`
	Data     string        `json:"data"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// Executor executes action plans using available tools
type Executor struct {
	availableTools []tools.Tool
	shortTerm      *memory.ShortTermMemory
	longTerm       *memory.LongTermMemory
}

// NewExecutor creates a new executor
func NewExecutor(availableTools []tools.Tool, shortTerm *memory.ShortTermMemory, longTerm *memory.LongTermMemory) *Executor {
	return &Executor{
		availableTools: availableTools,
		shortTerm:      shortTerm,
		longTerm:       longTerm,
	}
}

// Execute executes a plan
func (e *Executor) Execute(ctx context.Context, plan *Plan, userID string) (*ExecutionResult, error) {
	startTime := time.Now()

	result := &ExecutionResult{
		Output:      "",
		ToolResults: []ToolResult{},
		Context:     make(map[string]interface{}),
	}

	// Execute each action in order
	for i, action := range plan.Actions {
		log.Printf("Executing action %d: %s (%s)", i+1, action.Type, action.Tool)

		toolResult := e.executeAction(ctx, action, userID)
		result.ToolResults = append(result.ToolResults, toolResult)

		// If action failed, log but continue
		if toolResult.Error != "" {
			log.Printf("Action %d failed: %s", i+1, toolResult.Error)
		}

		// Add to context
		result.Context[action.Tool] = toolResult.Output
	}

	// Generate final output
	result.Output = e.generateOutput(result.ToolResults)

	result.Duration = time.Since(startTime)
	log.Printf("Plan execution completed in %v", result.Duration)

	return result, nil
}

func (e *Executor) executeAction(ctx context.Context, action Action, userID string) ToolResult {
	startTime := time.Now()

	// Handle non-execute actions
	if action.Type == "respond" {
		return ToolResult{
			Tool:     action.Tool,
			Output:   "Response generated",
			Data:     "Response generated",
			Duration: time.Since(startTime),
		}
	}

	// Find the tool
	var tool tools.Tool
	for _, t := range e.availableTools {
		if t.GetName() == action.Tool {
			tool = t
			break
		}
	}

	if tool == nil {
		return ToolResult{
			Tool:     action.Tool,
			Output:   "",
			Error:    fmt.Sprintf("Tool not found: %s", action.Tool),
			Duration: time.Since(startTime),
		}
	}

	// Execute the tool
	output, err := tool.Execute(ctx, action.Params)
	if err != nil {
		return ToolResult{
			Tool:     action.Tool,
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}
	}

	return ToolResult{
		Tool:     action.Tool,
		Output:   output,
		Data:     output,
		Duration: time.Since(startTime),
	}
}

func (e *Executor) generateOutput(results []ToolResult) string {
	if len(results) == 0 {
		return "No results to display."
	}

	var output string

	for _, result := range results {
		if result.Error != "" {
			output += fmt.Sprintf("\nError: %s", result.Error)
			continue
		}

		if result.Output != "" {
			output += result.Output + "\n"
		}
	}

	return output
}

// ExecuteSingle executes a single tool
func (e *Executor) ExecuteSingle(ctx context.Context, toolName string, params map[string]interface{}, userID string) (*ExecutionResult, error) {
	plan := &Plan{
		Actions: []Action{
			{
				Type:   "execute",
				Tool:   toolName,
				Params: params,
			},
		},
	}

	return e.Execute(ctx, plan, userID)
}
