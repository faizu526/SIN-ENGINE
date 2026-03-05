package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuggestRequest represents a suggestion request
type SuggestRequest struct {
	Query string `form:"q" json:"q"`
	Type  string `form:"type" json:"type"`
	Limit int    `form:"limit" json:"limit"`
}

// Suggestion represents a single suggestion
type Suggestion struct {
	Text  string  `json:"text"`
	Type  string  `json:"type"`
	Count int64   `json:"count"`
	Score float64 `json:"score"`
}

// TrendingTopic represents a trending topic
type TrendingTopic struct {
	Topic     string `json:"topic"`
	Searches  int64  `json:"searches"`
	Trend     string `json:"trend"` // "up", "down", "stable"
	Timestamp int64  `json:"timestamp"`
}

// RecentSearch represents a recent search
type RecentSearch struct {
	Query     string `json:"query"`
	Timestamp int64  `json:"timestamp"`
	UserID    string `json:"user_id,omitempty"`
}

// Suggest handles autocomplete suggestions
func Suggest(queryEngine interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query is required"})
			return
		}

		limit := parseInt(c.Query("limit"), 10)
		if limit > 20 {
			limit = 20
		}

		// TODO: Use queryEngine.Suggest() when implemented
		// For now, return mock suggestions
		suggestions := []Suggestion{
			{Text: query + " tutorial", Type: "search", Count: 100, Score: 0.9},
			{Text: query + " course", Type: "course", Count: 50, Score: 0.8},
			{Text: query + " github", Type: "resource", Count: 30, Score: 0.7},
		}

		c.JSON(http.StatusOK, gin.H{
			"query":       query,
			"suggestions": suggestions,
		})
	}
}

// Trending returns trending search topics
func Trending() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := parseInt(c.Query("limit"), 10)
		if limit > 20 {
			limit = 20
		}

		// TODO: Fetch from Redis/cache
		trending := []TrendingTopic{
			{Topic: "Penetration Testing", Searches: 1500, Trend: "up", Timestamp: 1704067200},
			{Topic: "SQL Injection", Searches: 1200, Trend: "up", Timestamp: 1704067200},
			{Topic: "XSS Prevention", Searches: 900, Trend: "stable", Timestamp: 1704067200},
			{Topic: "Burp Suite", Searches: 800, Trend: "up", Timestamp: 1704067200},
			{Topic: "Metasploit", Searches: 750, Trend: "down", Timestamp: 1704067200},
		}

		c.JSON(http.StatusOK, gin.H{
			"trending": trending,
		})
	}
}

// Recent returns recent searches
func Recent() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := parseInt(c.Query("limit"), 10)
		if limit > 50 {
			limit = 50
		}

		// TODO: Fetch from Redis
		recent := []RecentSearch{
			{Query: "how to hack wifi", Timestamp: 1704123456},
			{Query: "sql injection examples", Timestamp: 1704112345},
			{Query: "xss payload list", Timestamp: 1704101234},
		}

		c.JSON(http.StatusOK, gin.H{
			"recent": recent,
		})
	}
}
