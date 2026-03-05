package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/ai-assistant/agents"
)

// CommandRequest represents a command request
type CommandRequest struct {
	Command string                 `json:"command" binding:"required"`
	Params  map[string]interface{} `json:"params"`
	UserID  string                 `json:"user_id"`
}

// ExecuteCommand handles command execution
func ExecuteCommand(executor *agents.Executor) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CommandRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.UserID == "" {
			req.UserID = "anonymous"
		}

		// Create a simple plan with just this command
		plan := &agents.Plan{
			Actions: []agents.Action{
				{
					Type:   "execute",
					Tool:   req.Command,
					Params: req.Params,
				},
			},
		}

		// Execute
		result, err := executor.Execute(c.Request.Context(), plan, req.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"command": req.Command,
			"result":  result.Output,
			"success": len(result.ToolResults) > 0 && result.ToolResults[0].Error == nil,
			"took":    result.Duration,
		})
	}
}

// ListCommands returns available commands
func ListCommands() gin.HandlerFunc {
	return func(c *gin.Context) {
		commands := []map[string]interface{}{
			{
				"name":        "search",
				"description": "Search for cybersecurity resources",
				"params":      []string{"query", "category"},
			},
			{
				"name":        "scan",
				"description": "Scan a target for vulnerabilities",
				"params":      []string{"target", "scan_type"},
			},
			{
				"name":        "dork",
				"description": "Search using Google dorks",
				"params":      []string{"dork", "engine"},
			},
			{
				"name":        "breach_check",
				"description": "Check if email/username in breach",
				"params":      []string{"email", "username"},
			},
			{
				"name":        "crawl",
				"description": "Crawl a website",
				"params":      []string{"url", "depth"},
			},
			{
				"name":        "explain",
				"description": "Explain a security concept",
				"params":      []string{"topic"},
			},
		}

		c.JSON(http.StatusOK, gin.H{"commands": commands})
	}
}
