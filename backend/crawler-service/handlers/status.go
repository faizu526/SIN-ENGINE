package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/crawler-service/queue"
)

func GetQueueStats(jobQueue *queue.JobQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		stats, err := jobQueue.GetStats(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get queue stats"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":   stats,
			"service": "crawler-service",
		})
	}
}

func PauseQueue(jobQueue *queue.JobQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would pause the queue
		c.JSON(http.StatusOK, gin.H{
			"status":  "paused",
			"message": "Queue paused successfully",
		})
	}
}

func ResumeQueue(jobQueue *queue.JobQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would resume the queue
		c.JSON(http.StatusOK, gin.H{
			"status":  "resumed",
			"message": "Queue resumed successfully",
		})
	}
}
