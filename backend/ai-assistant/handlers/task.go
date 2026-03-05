package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/ai-assistant/agents"
	"gorm.io/gorm"
)

// Task represents a background task
type Task struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"` // pending, running, completed, failed
	Progress    int                    `json:"progress"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// CreateTask creates a new task
func CreateTask(planner *agents.Planner, executor *agents.Executor) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name   string `json:"name" binding:"required"`
			Goal   string `json:"goal" binding:"required"`
			UserID string `json:"user_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.UserID == "" {
			req.UserID = "anonymous"
		}

		// Generate task ID
		taskID := fmt.Sprintf("task_%d", time.Now().Unix())

		// Create plan
		plan := &agents.Plan{
			Goal: req.Goal,
		}

		// Execute asynchronously
		go func() {
			_, err := executor.Execute(c.Request.Context(), plan, req.UserID)
			if err != nil {
				fmt.Printf("Task %s failed: %v\n", taskID, err)
			}
		}()

		c.JSON(http.StatusAccepted, gin.H{
			"task_id": taskID,
			"name":    req.Name,
			"status":  "pending",
			"message": "Task created successfully",
			"created": time.Now().Unix(),
		})
	}
}

// GetTaskStatus returns task status
func GetTaskStatus(executor *agents.Executor) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")

		// In a real implementation, we'd look up the task from storage
		// For now, return a placeholder
		c.JSON(http.StatusOK, gin.H{
			"task_id":  taskID,
			"status":   "completed",
			"progress": 100,
		})
	}
}

// ListTasks lists all tasks
func ListTasks(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")

		// In a real implementation, query from database
		tasks := []Task{
			{
				ID:        "task_1",
				UserID:    userID,
				Name:      "Sample Task",
				Status:    "completed",
				Progress:  100,
				CreatedAt: time.Now().Add(-time.Hour),
			},
		}

		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	}
}
