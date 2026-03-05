package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/ai-assistant/agents"
	"github.com/sin-engine/ai-assistant/memory"
)

// ChatRequest represents a chat request
type ChatRequest struct {
	Message   string                 `json:"message" binding:"required"`
	UserID    string                 `json:"user_id"`
	Context   map[string]interface{} `json:"context"`
	Stream    bool                   `json:"stream"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Response   string                 `json:"response"`
	Intent     string                 `json:"intent"`
	Confidence float64                `json:"confidence"`
	Actions    []string               `json:"actions,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
}

// Chat handles chat requests
func Chat(
	analyzer *agents.Analyzer,
	planner *agents.Planner,
	executor *agents.Executor,
	shortTerm *memory.ShortTermMemory,
	longTerm *memory.LongTermMemory,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		// Set default user ID if not provided
		if req.UserID == "" {
			req.UserID = "anonymous"
		}

		ctx := c.Request.Context()

		// Step 1: Analyze the message
		analysis, err := analyzer.Analyze(ctx, req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to analyze message: " + err.Error(),
			})
			return
		}

		log.Printf("User %s: %s (Intent: %s, Confidence: %.2f)", 
			req.UserID, req.Message, analysis.Intent, analysis.Confidence)

		// Step 2: Get conversation context from short-term memory
		history, _ := shortTerm.GetRecent(ctx, req.UserID, 10)

		// Step 3: Get relevant knowledge from long-term memory
		knowledge, _ := longTerm.Search(ctx, req.Message, 3)

		// Step 4: Plan actions
		plan, err := planner.CreatePlan(ctx, analysis, knowledge, req.Context)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create plan: " + err.Error(),
			})
			return
		}

		// Step 5: Execute plan
		result, err := executor.Execute(ctx, plan, req.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to execute: " + err.Error(),
			})
			return
		}

		// Step 6: Generate response
		response := generateResponse(analysis, plan, result, knowledge)

		// Step 7: Store in short-term memory
		shortTerm.Add(ctx, req.UserID, memory.Message{
			Role:    "user",
			Content: req.Message,
			Time:    time.Now(),
		})

		shortTerm.Add(ctx, req.UserID, memory.Message{
			Role:    "assistant",
			Content: response,
			Time:    time.Now(),
		})

		// Step 8: Learn from interaction if important
		if analysis.Confidence > 0.8 && isLearningOpportunity(analysis.Intent) {
			longTerm.Store(ctx, req.UserID, memory.KnowledgeItem{
				Topic:    analysis.Intent,
				Content:  fmt.Sprintf("User asked about: %s, Response: %s", req.Message, response),
				Source:   "conversation",
				Tags:     []string{analysis.Intent},
				CreatedAt: time.Now(),
			})
		}

		c.JSON(http.StatusOK, ChatResponse{
			Response:   response,
			Intent:     analysis.Intent,
			Confidence: analysis.Confidence,
			Actions:    plan.Actions,
			Context:    result.Context,
			Timestamp:  time.Now().Unix(),
		})
	}
}

// ChatStream handles streaming chat
func ChatStream(
	analyzer *agents.Analyzer,
	planner *agents.Planner,
	executor *agents.Executor,
	shortTerm *memory.ShortTermMemory,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.UserID == "" {
			req.UserID = "anonymous"
		}

		// Set headers for streaming
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		// Analyze
		analysis, _ := analyzer.Analyze(c.Request.Context(), req.Message)

		// Create plan
		plan, _ := planner.CreatePlan(c.Request.Context(), analysis, nil, req.Context)

		// Stream response
		responseChan := make(chan string)
		errChan := make(chan error)

		go func() {
			result, err := executor.Execute(c.Request.Context(), plan, req.UserID)
			if err != nil {
				errChan <- err
				return
			}
			response := generateResponse(analysis, plan, result, nil)
			responseChan <- response
		}()

		c.Stream(func(w io.Writer) bool {
			select {
			case err := <-errChan:
				c.SSEvent("error", gin.H{"error": err.Error()})
				return false
			case response := <-responseChan:
				c.SSEvent("message", response)
				return false
			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}

// generateResponse generates a natural language response
func generateResponse(analysis *agents.Analysis, plan *agents.Plan, result *agents.ExecutionResult, knowledge []memory.KnowledgeItem) string {
	if len(result.ToolResults) == 0 {
		// No tools used, direct response
		return generateDirectResponse(analysis.Intent, analysis.Entities)
	}

	// Generate response based on tool results
	var response string

	switch analysis.Intent {
	case "search":
		response = formatSearchResults(result.ToolResults)
	case "scan":
		response = formatScanResults(result.ToolResults)
	case "exploit":
		response = formatExploitResults(result.ToolResults)
	case "breach_check":
		response = formatBreachResults(result.ToolResults)
	case "learn":
		response = formatLearningResults(knowledge)
	default:
		response = generateDirectResponse(analysis.Intent, analysis.Entities)
	}

	return response
}

func generateDirectResponse(intent string, entities map[string]string) string {
	responses := map[string]string{
		"greeting":    "Hello! I'm your cybersecurity learning assistant. How can I help you today?",
		"help":        "I can help you with: searching for learning resources, scanning targets, checking breaches, running exploits, and more!",
		"explain":     "I can explain various cybersecurity concepts. What would you like to learn about?",
		"search":      "I'll search for relevant learning resources for you.",
		"scan":        "I'll run a scan on the target you specified.",
		"breach_check": "I'll check if your credentials have been in any data breaches.",
		"dork":        "I'll search using Google dorks to find relevant information.",
		"learn":       "Great! Let me find the best learning resources for you.",
	}

	if resp, ok := responses[intent]; ok {
		return resp
	}

	return "I understand you're interested in " + intent + ". Let me help you with that!"
}

func formatSearchResults(results []agents.ToolResult) string {
	if len(results) == 0 {
		return "No search results found."
	}

	response := "I found the following resources:\n"
	for i, r := range results {
		if i >= 5 {
			break
		}
		response += fmt.Sprintf("\n%d. %s", i+1, r.Data)
	}

	return response
}

func formatScanResults(results []agents.ToolResult) string {
	if len(results) == 0 {
	 results found."
		return "No scan}

	response := "Scan completed! Here are the findings:\n"
	for _, r := range results {
		response += fmt.Sprintf("\n- %s", r.Data)
	}

	return response
}

func formatExploitResults(results []agents.ToolResult) string {
	if len(results) == 0 {
		return "No exploit results found."
	}

	return "Exploit search completed! Found the following:\n" + results[0].Data
}

func formatBreachResults(results []agents.ToolResult) string {
	if len(results) == 0 {
		return "No breach information found. Your credentials appear safe!"
	}

	return "Warning! Found breach information:\n" + results[0].Data
}

func formatLearningResults(knowledge []memory.KnowledgeItem) string {
	if len(knowledge) == 0 {
		return "I don't have specific knowledge about that topic yet. Would you like me to search for learning resources?"
	}

	response := "Based on what you've learned:\n"
	for _, k := range knowledge {
		response += fmt.Sprintf("\n- %s: %s", k.Topic, k.Content)
	}

	return response
}

func isLearningOpportunity(intent string) bool {
	learningIntents := map[string]bool{
		"explain":   true,
		"learn":     true,
		"search":    true,
		"breach_check": true,
	}
	return learningIntents[intent]
}
