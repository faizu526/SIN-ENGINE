package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sin-engine/dork-service/engines"
	"gorm.io/gorm"
)

// DorkRequest represents a dork search request
type DorkRequest struct {
	Query      string   `json:"query" binding:"required"`
	Engines    []string `json:"engines"`
	Limit      int      `json:"limit"`
	Category   string   `json:"category"`
	CacheTTL   int      `json:"cache_ttl"`
}

// SearchDork handles multi-engine dork search
func SearchDork(
	google *engines.GoogleEngine,
	shodan *engines.ShodanEngine,
	censys *engines.CensysEngine,
	fofa *engines.FOFAEngine,
	github *engines.GitHubEngine,
	redisClient *redis.Client,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DorkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		// Set defaults
		if req.Limit <= 0 {
			req.Limit = 10
		}
		if req.CacheTTL <= 0 {
			req.CacheTTL = 3600 // 1 hour default
		}
		if len(req.Engines) == 0 {
			req.Engines = []string{"google", "shodan", "github"}
		}

		// Determine category from query if not specified
		if req.Category == "" {
			req.Category = detectCategory(req.Query)
		}

		// Results channel
		type result struct {
			engine  string
			results []engines.SearchResult
			err     error
		}

		resultsChan := make(chan result, len(req.Engines))

		// Run searches in parallel
		for _, engine := range req.Engines {
			go func(eng string) {
				var results []engines.SearchResult
				var err error

				ttl := time.Duration(req.CacheTTL) * time.Second

				switch eng {
				case "google":
					results, err = google.SearchWithCache(c.Request.Context(), redisClient, req.Query, req.Limit, ttl)
				case "shodan":
					results, err = shodan.SearchWithCache(c.Request.Context(), redisClient, req.Query, req.Limit, ttl)
				case "censys":
					results, err = censys.SearchWithCache(c.Request.Context(), redisClient, req.Query, req.Limit, ttl)
				case "fofa":
					results, err = fofa.SearchWithCache(c.Request.Context(), redisClient, req.Query, req.Limit, ttl)
				case "github":
					results, err = github.SearchWithCache(c.Request.Context(), redisClient, req.Query, req.Limit, ttl)
				default:
					err = fmt.Errorf("unknown engine: %s", eng)
				}

				resultsChan <- result{engine: eng, results: results, err: err}
			}(engine)
		}

		// Collect results
		allResults := make([]engines.SearchResult, 0)
		errors := make(map[string]string)

		for i := 0; i < len(req.Engines); i++ {
			res := <-resultsChan
			if res.err != nil {
				errors[res.engine] = res.err.Error()
				log.Printf("Engine %s error: %v", res.engine, res.err)
				continue
			}

			// Add category to results
			for i := range res.results {
				if res.results[i].Category == "" {
					res.results[i].Category = req.Category
				}
			}

			allResults = append(allResults, res.results...)
			log.Printf("Engine %s returned %d results", res.engine, len(res.results))
		}

		// Limit total results
		if len(allResults) > req.Limit*len(req.Engines) {
			allResults = allResults[:req.Limit*len(req.Engines)]
		}

		c.JSON(http.StatusOK, gin.H{
			"query":     req.Query,
			"category":  req.Category,
			"total":     len(allResults),
			"results":   allResults,
			"engines":   req.Engines,
			"errors":    errors,
			"timestamp": time.Now().Unix(),
		})
	}
}

// SearchGoogle handles Google dork search
func SearchGoogle(engine *engines.GoogleEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		limit := 10
		if l := c.Query("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}

		results, err := engine.Search(c.Request.Context(), query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":    query,
			"engine":   "google",
			"total":    len(results),
			"results":  results,
		})
	}
}

// SearchShodan handles Shodan search
func SearchShodan(engine *engines.ShodanEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		limit := 10
		if l := c.Query("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}

		results, err := engine.Search(c.Request.Context(), query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":    query,
			"engine":   "shodan",
			"total":    len(results),
			"results":  results,
		})
	}
}

// SearchCensys handles Censys search
func SearchCensys(engine *engines.CensysEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		limit := 10
		if l := c.Query("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}

		results, err := engine.Search(c.Request.Context(), query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":    query,
			"engine":   "censys",
			"total":    len(results),
			"results":  results,
		})
	}
}

// SearchFOFA handles FOFA search
func SearchFOFA(engine *engines.FOFAEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		limit := 10
		if l := c.Query("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}

		results, err := engine.Search(c.Request.Context(), query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":    query,
			"engine":   "fofa",
			"total":    len(results),
			"results":  results,
		})
	}
}

// SearchGitHub handles GitHub search
func SearchGitHub(engine *engines.GitHubEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		limit := 10
		if l := c.Query("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}

		results, err := engine.Search(c.Request.Context(), query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":    query,
			"engine":   "github",
			"total":    len(results),
			"results":  results,
		})
	}
}

// GetCategories returns available dork categories
func GetCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		categories := []map[string]interface{}{
			{"id": "vulnerability", "name": "Vulnerability", "description": "Find vulnerable systems"},
			{"id": "exploit", "name": "Exploit", "description": "Find exploits and PoCs"},
			{"id": "config", "name": "Configuration", "description": "Misconfigured systems"},
			{"id": "sensitive", "name": "Sensitive Data", "description": "Exposed sensitive information"},
			{"id": "credentials", "name": "Credentials", "description": "Leaked credentials"},
			{"id": "iot", "name": "IoT", "description": "Internet of Things devices"},
			{"id": "database", "name": "Database", "description": "Exposed databases"},
			{"id": "admin", "name": "Admin Panels", "description": "Admin login pages"},
			{"id": "camera", "name": "Cameras", "description": "IP cameras"},
			{"id": "server", "name": "Servers", "description": "Server technologies"},
		}

		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

// GetPopularDorks returns popular dork queries
func GetPopularDorks(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try Redis first
		if redisClient != nil {
			popular, err := redisClient.ZRevRangeWithScores(c.Request.Context(), "dork:popular", 0, 9).Result()
			if err == nil && len(popular) > 0 {
				dorks := make([]map[string]interface{}, len(popular))
				for i, item := range popular {
					dorks[i] = map[string]interface{}{
						"query": item.Member,
						"count": int(item.Score),
					}
				}
				c.JSON(http.StatusOK, gin.H{"dorks": dorks})
				return
			}
		}

		// Fallback to default popular dorks
		popularDorks := []string{
			"site:github.com password",
			"inurl:admin login",
			"filetype:sql mysql",
			"intitle:index.of /config",
			"inurl:wp-content/wp-config.php",
			"site:*.edu vulnerable",
			"inurl:cgi-bin admin",
			"filetype:pem private key",
			"intitle:phpinfo",
			"inurl:/phpmyadmin",
		}

		c.JSON(http.StatusOK, gin.H{"dorks": popularDorks})
	}
}

// SaveDork saves a dork to history
func SaveDork(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Query    string `json:"query" binding:"required"`
			Engine   string `json:"engine"`
			UserID   uint   `json:"user_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if db == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Dork saved (demo mode)"})
			return
		}

		// Note: In production, create a DorkHistory model
		c.JSON(http.StatusOK, gin.H{
			"message": "Dork saved successfully",
			"query":   req.Query,
		})
	}
}

// GetDorkHistory returns user's dork search history
func GetDorkHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusOK, gin.H{"history": []})
			return
		}

		// Note: In production, query from DorkHistory model
		c.JSON(http.StatusOK, gin.H{"history": []})
	}
}

// detectCategory determines the category from a query
func detectCategory(query string) string {
	lowerQuery := query

	// Simple keyword matching
	categories := map[string][]string{
		"vulnerability": {"vuln", "vulnerable", "cve", "exploit", "poc"},
		"config":         {"config", "conf", "setup", "install"},
		"sensitive":      {"sensitive", "private", "confidential", "secret"},
		"credentials":    {"password", "credential", "username", "login", "passwd"},
		"iot":           {"iot", "camera", "webcam", "smart", "device"},
		"database":      {"database", "db", "mysql", "mongodb", "redis"},
		"admin":         {"admin", "panel", "dashboard", "login"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if len(lowerQuery) > len(keyword) && (lowerQuery[:len(keyword)] == keyword || len(lowerQuery) > 5 && contains(lowerQuery, keyword)) {
				return category
			}
		}
	}

	return "general"
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
