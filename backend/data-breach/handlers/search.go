package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/data-breach/analyzers"
	"github.com/sin-engine/data-breach/collectors"
	"gorm.io/gorm"
)

// Search handles breach search requests
func Search(
	collectorsList []collectors.Collector,
	parser *analyzers.Parser,
	db *gorm.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Query string `json:"query" binding:"required"`
			Types string `json:"types"` // email, domain, username
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: query is required",
			})
			return
		}

		// Validate query
		validator := analyzers.NewValidator()
		if err := validator.ValidateSearchQuery(request.Query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set default search type
		if request.Types == "" {
			request.Types = "email"
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		// Search all collectors
		var allBreaches []collectors.Breach
		sources := make([]string, 0)

		for _, collector := range collectorsList {
			if !collector.IsAvailable() {
				continue
			}

			breaches, err := collector.Search(ctx, request.Query)
			if err != nil {
				continue
			}

			allBreaches = append(allBreaches, breaches...)
			sources = append(sources, collector.GetName())
		}

		// Calculate risk score
		var dataClasses []string
		for _, breach := range allBreaches {
			dataClasses = append(dataClasses, breach.DataClasses...)
		}
		riskScore := validator.CalculateRiskScore(dataClasses)

		// Build response
		response := gin.H{
			"query":      request.Query,
			"breaches":   allBreaches,
			"sources":    sources,
			"total":      len(allBreaches),
			"risk_score": riskScore,
			"timestamp":  time.Now().Unix(),
		}

		// Store search result in database if DB is available
		if db != nil {
			// TODO: Save to database
			_ = db
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetSearchResult retrieves a previous search result
func GetSearchResult(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Param("email")

		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database not available",
			})
			return
		}

		// TODO: Query database for previous search
		_ = email

		c.JSON(http.StatusOK, gin.H{
			"message": "Search result retrieval not yet implemented",
			"email":   email,
		})
	}
}
