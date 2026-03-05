package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sin-engine/crawler-service/engine"
	"github.com/sin-engine/crawler-service/queue"
)

type StartCrawlRequest struct {
	URL        string   `json:"url" binding:"required"`
	Depth      int      `json:"depth"`
	MaxDepth   int      `json:"max_depth"`
	UserAgent  string   `json:"user_agent"`
	Concurrent int      `json:"concurrent"`
	Timeout    int      `json:"timeout"`
	Filters    []string `json:"filters"`
	Priority   int      `json:"priority"`
	Schedule   string   `json:"schedule"`
}

type StartCrawlResponse struct {
	JobID       string `json:"job_id"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	ScheduledAt int64  `json:"scheduled_at,omitempty"`
}

func StartCrawl(crawler *engine.CrawlerEngine, jobQueue *queue.JobQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StartCrawlRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Validate URL
		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
			return
		}

		// Set defaults
		if req.MaxDepth == 0 {
			req.MaxDepth = 3
		}
		if req.Timeout == 0 {
			req.Timeout = 30
		}
		if req.Concurrent == 0 {
			req.Concurrent = 5
		}

		jobID := uuid.New().String()

		// Check if scheduled for later
		if req.Schedule != "" {
			scheduledTime, err := time.Parse(time.RFC3339, req.Schedule)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule time format"})
				return
			}

			// Add to queue with scheduled time
			job := queue.Job{
				ID:          jobID,
				URL:         req.URL,
				Depth:       0,
				MaxDepth:    req.MaxDepth,
				UserAgent:   req.UserAgent,
				Concurrent:  req.Concurrent,
				Timeout:     req.Timeout,
				Filters:     req.Filters,
				Priority:    req.Priority,
				ScheduledAt: scheduledTime,
			}

			if err := jobQueue.Push(c.Request.Context(), job); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule job"})
				return
			}

			c.JSON(http.StatusAccepted, StartCrawlResponse{
				JobID:       jobID,
				URL:         req.URL,
				Status:      "scheduled",
				ScheduledAt: scheduledTime.Unix(),
			})
			return
		}

		// Create and start job immediately
		job := &engine.CrawlJob{
			ID:         jobID,
			URL:        req.URL,
			Depth:      0,
			MaxDepth:   req.MaxDepth,
			UserAgent:  req.UserAgent,
			Concurrent: req.Concurrent,
			Timeout:    req.Timeout,
			Filters:    req.Filters,
			Status:     "pending",
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(req.Timeout)*time.Second)
		defer cancel()

		if err := crawler.StartJob(ctx, job); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start crawl: " + err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, StartCrawlResponse{
			JobID:  jobID,
			URL:    req.URL,
			Status: "started",
		})
	}
}

func StopCrawl(crawler *engine.CrawlerEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("job_id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
			return
		}

		if err := crawler.StopJob(jobID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"job_id": jobID,
			"status": "stopped",
		})
	}
}

func GetCrawlStatus(crawler *engine.CrawlerEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("job_id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
			return
		}

		job := crawler.GetJobStatus(jobID)
		if job == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"job_id":     job.ID,
			"url":        job.URL,
			"status":     job.Status,
			"results":    len(job.Results),
			"start_time": job.StartTime.Unix(),
			"end_time":   job.EndTime.Unix(),
		})
	}
}

func GetCrawlResults(store interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("job_id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
			return
		}

		// This would fetch results from storage
		// For now, return placeholder
		c.JSON(http.StatusOK, gin.H{
			"job_id":  jobID,
			"results": []interface{}{},
			"total":   0,
		})
	}
}
