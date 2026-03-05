package querier

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// QueryEngine handles search queries
type QueryEngine struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// SearchQuery represents a search query
type SearchQuery struct {
	Text       string
	Type       string
	Page       int
	Limit      int
	Category   string
	Difficulty string
	SortBy     string
}

// SearchResult represents a search result
type SearchResult struct {
	ID          string
	Type        string
	Title       string
	Description string
	URL         string
	Image       string
	Score       float64
	Metadata    map[string]interface{}
	Timestamp   time.Time
}

// NewQueryEngine creates a new query engine
func NewQueryEngine(db *gorm.DB, redisClient *redis.Client) *QueryEngine {
	return &QueryEngine{
		db:          db,
		redisClient: redisClient,
	}
}

// Search performs a search query
func (q *QueryEngine) Search(query SearchQuery) ([]SearchResult, int64, error) {
	ctx := context.Background()

	// Build search terms from query
	terms := strings.Fields(strings.ToLower(query.Text))

	// Log the search query for analytics
	q.logSearch(query.Text, terms)

	var results []SearchResult

	// Search in Redis cache first
	if q.redisClient != nil {
		results = q.searchCache(ctx, terms, query)
		if len(results) > 0 {
			return results, int64(len(results)), nil
		}
	}

	// Fall back to database search
	results, err := q.searchDatabase(query)
	if err != nil {
		return nil, 0, err
	}

	// Cache results
	if q.redisClient != nil && len(results) > 0 {
		q.cacheResults(ctx, query, results)
	}

	return results, int64(len(results)), nil
}

// searchCache searches in Redis cache
func (q *QueryEngine) searchCache(ctx context.Context, terms []string, query SearchQuery) []SearchResult {
	var results []SearchResult

	// Build cache key
	cacheKey := "search:results:" + strings.Join(terms, ":") + ":" + query.Type

	// This would retrieve cached results
	// For now, return empty to trigger database search

	return results
}

// searchDatabase searches in PostgreSQL
func (q *QueryEngine) searchDatabase(query SearchQuery) ([]SearchResult, error) {
	var results []SearchResult

	// This would build a proper SQL query based on the search terms
	// For demonstration, we'll create some sample results

	if query.Text != "" {
		// Simulate search results
		results = []SearchResult{
			{
				ID:          "1",
				Type:        "course",
				Title:       "Ethical Hacking Complete Course",
				Description: "Learn ethical hacking from scratch",
				URL:         "/courses/ethical-hacking",
				Image:       "/images/courses/ethical-hacking.jpg",
				Score:       0.95,
				Timestamp:   time.Now(),
			},
			{
				ID:          "2",
				Type:        "platform",
				Title:       "TryHackMe",
				Description: "Learn cybersecurity through hands-on labs",
				URL:         "/platforms/tryhackme",
				Image:       "/images/platforms/tryhackme.png",
				Score:       0.90,
				Timestamp:   time.Now(),
			},
			{
				ID:          "3",
				Type:        "tool",
				Title:       "Burp Suite",
				Description: "Web application security testing tool",
				URL:         "/tools/burp-suite",
				Image:       "/images/tools/burp-suite.png",
				Score:       0.85,
				Timestamp:   time.Now(),
			},
		}
	}

	return results, nil
}

// cacheResults caches search results in Redis
func (q *QueryEngine) cacheResults(ctx context.Context, query SearchQuery, results []SearchResult) {
	// This would serialize and cache results with TTL
}

// Suggest returns autocomplete suggestions
func (q *QueryEngine) Suggest(query string, limit int) ([]string, error) {
	// This would use Redis sorted sets for prefix matching
	suggestions := []string{
		query + " tutorial",
		query + " course",
		query + " github",
		query + " pdf",
		query + " writeup",
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions, nil
}

// logSearch logs search queries for analytics
func (q *QueryEngine) logSearch(query string, terms []string) {
	log.Printf("Search: %s (terms: %v)", query, terms)

	// Increment search count in Redis
	if q.redisClient != nil {
		ctx := context.Background()
		key := "search:stats:queries"
		q.redisClient.Incr(ctx, key)

		// Also track individual terms
		for _, term := range terms {
			q.redisClient.ZIncrBy(ctx, "search:stats:terms", 1, term)
		}
	}
}
