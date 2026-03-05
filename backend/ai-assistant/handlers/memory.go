package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/ai-assistant/memory"
)

// GetShortTermMemory returns short-term memory for a user
func GetShortTermMemory(shortTerm *memory.ShortTermMemory) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		messages, err := shortTerm.GetRecent(c.Request.Context(), userID, 50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"messages": messages})
	}
}

// GetLongTermMemory returns long-term memory for a user
func GetLongTermMemory(longTerm *memory.LongTermMemory) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		items, err := longTerm.GetAll(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"knowledge": items})
	}
}

// ClearMemory clears memory for a user
func ClearMemory(shortTerm *memory.ShortTermMemory, longTerm *memory.LongTermMemory) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		shortTerm.Clear(c.Request.Context(), userID)
		longTerm.Clear(c.Request.Context(), userID)

		c.JSON(http.StatusOK, gin.H{"message": "Memory cleared"})
	}
}

// ListTools returns available tools
func ListTools(tools []memory.Tool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tools": tools})
	}
}

// ExecuteTool executes a specific tool
func ExecuteTool(tools []memory.Tool) gin.HandlerFunc {
	return func(c *gin.Context) {
		toolName := c.Param("name")
		c.JSON(http.StatusOK, gin.H{"message": "Tool execution not implemented"})
	}
}

// Learn stores new knowledge
func Learn(db interface{}, longTerm *memory.LongTermMemory) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Learning not implemented"})
	}
}

// GetKnowledge retrieves knowledge about a topic
func GetKnowledge(longTerm *memory.LongTermMemory) gin.HandlerFunc {
	return func(c *gin.Context) {
		topic := c.Param("topic")
		userID := c.Query("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		item, err := longTerm.Get(c.Request.Context(), userID, topic)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Knowledge not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"knowledge": item})
	}
}
