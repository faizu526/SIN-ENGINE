package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/search-service/querier"
	"github.com/sin-engine/search-service/ranker"
)

type SearchRequest struct {
	Query      string `form:"q" json:"q"`
	Type       string `form:"type" json:"type"`
	Page       int    `form:"page" json:"page"`
	Limit      int    `form:"limit" json:"limit"`
	Category   string `form:"category" json:"category"`
	Difficulty string `form:"difficulty" json:"difficulty"`
	SortBy     string `form:"sort" json:"sort"`
}

type SearchResult struct {
	ID          string                 `json:"id"`
	ResultType  string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
	Image       string                 `json:"image,omitempty"`
	Score       float64                `json:"score"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
	Query   string         `json:"query"`
	Time    float64        `json:"time_ms"`
}

// Search handler - main search endpoint
func Search(qe *querier.QueryEngine, re *ranker.RankerEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Parse request
		req := SearchRequest{
			Query:      c.Query("q"),
			Type:       c.DefaultQuery("type", "all"),
			Page:       parseInt(c.Query("page"), 1),
			Limit:      parseInt(c.Query("limit"), 20),
			Category:   c.Query("category"),
			Difficulty: c.Query("difficulty"),
			SortBy:     c.DefaultQuery("sort", "relevance"),
		}

		// Validate query
		if strings.TrimSpace(req.Query) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query is required"})
			return
		}

		// Limit results
		if req.Limit > 100 {
			req.Limit = 100
		}
		if req.Limit < 1 {
			req.Limit = 20
		}

		// Build search query
		query := querier.SearchQuery{
			Text:       req.Query,
			Type:       req.Type,
			Page:       req.Page,
			Limit:      req.Limit,
			Category:   req.Category,
			Difficulty: req.Difficulty,
			SortBy:     req.SortBy,
		}

		// Execute search
		results, total, err := qe.Search(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed: " + err.Error()})
			return
		}

		// Re-rank results if needed
		if req.SortBy == "relevance" && len(results) > 0 {
			rankerResults := convertToRankerResults(results)
			rankerResults, err = re.Rerank(rankerResults, req.Query)
			if err != nil {
				log.Printf("Rerank error: %v", err)
			}
			results = convertFromRankerResults(rankerResults)
		}

		// Convert to response format
		searchResults := make([]SearchResult, len(results))
		for i, r := range results {
			searchResults[i] = SearchResult{
				ID:          r.ID,
				ResultType:  r.Type,
				Title:       r.Title,
				Description: r.Description,
				URL:         r.URL,
				Image:       r.Image,
				Score:       r.Score,
				Metadata:    r.Metadata,
				Timestamp:   r.Timestamp,
			}
		}

		elapsed := time.Since(start).Milliseconds()

		c.JSON(http.StatusOK, SearchResponse{
			Results: searchResults,
			Total:   total,
			Page:    req.Page,
			Limit:   req.Limit,
			Query:   req.Query,
			Time:    float64(elapsed),
		})
	}
}

// Convert querier results to ranker results
func convertToRankerResults(results []querier.SearchResult) []ranker.SearchResultItem {
	rankerResults := make([]ranker.SearchResultItem, len(results))
	for i, r := range results {
		rankerResults[i] = ranker.SearchResultItem{
			ID:          r.ID,
			Type:        r.Type,
			Title:       r.Title,
			Description: r.Description,
			URL:         r.URL,
			Image:       r.Image,
			Score:       r.Score,
			Metadata:    r.Metadata,
		}
	}
	return rankerResults
}

// Convert ranker results back to querier results
func convertFromRankerResults(results []ranker.SearchResultItem) []querier.SearchResult {
	querierResults := make([]querier.SearchResult, len(results))
	for i, r := range results {
		querierResults[i] = querier.SearchResult{
			ID:          r.ID,
			Type:        r.Type,
			Title:       r.Title,
			Description: r.Description,
			URL:         r.URL,
			Image:       r.Image,
			Score:       r.Score,
			Metadata:    r.Metadata,
			Timestamp:   time.Now(),
		}
	}
	return querierResults
}

// Suggest handler - autocomplete suggestions
func Suggest(qe *querier.QueryEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if strings.TrimSpace(query) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query is required"})
			return
		}

		limit := parseInt(c.Query("limit"), 10)

		suggestions, err := qe.Suggest(query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Suggest failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":       query,
			"suggestions": suggestions,
		})
	}
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
