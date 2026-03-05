package engines

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/redis/go-redis/v9"
)

// GoogleEngine handles Google dork searches
type GoogleEngine struct {
	client  *http.Client
	baseURL string
}

// NewGoogleEngine creates a new Google search engine
func NewGoogleEngine() *GoogleEngine {
	return &GoogleEngine{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		baseURL: "https://www.google.com/search",
	}
}

// SearchResult represents a search result
type SearchResult struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Snippet   string `json:"snippet"`
	Engine    string `json:"engine"`
	Category  string `json:"category,omitempty"`
	IndexedAt int64  `json:"indexed_at"`
}

// Search performs a Google dork search
func (e *GoogleEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	results := make([]SearchResult, 0, limit)

	// Build search URL with dork
	searchURL := fmt.Sprintf("%s?q=%s&num=%d", e.baseURL, url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract results
	doc.Find(".g").Each(func(i int, s *goquery.Selection) {
		if len(results) >= limit {
			return
		}

		// Find title and link
		title := s.Find("h3").First().Text()
		linkElem := s.Find("a").First()
		link, _ := linkElem.Attr("href")

		// Skip if no valid link
		if link == "" || len(link) < 10 {
			return
		}

		// Extract snippet
		snippet := s.Find(".VwiC3b").First().Text()
		if snippet == "" {
			snippet = s.Find("span").First().Text()
		}

		results = append(results, SearchResult{
			Title:     title,
			URL:       link,
			Snippet:   snippet,
			Engine:    "google",
			IndexedAt: time.Now().Unix(),
		})
	})

	log.Printf("Google search for '%s' returned %d results", query, len(results))
	return results, nil
}

// SearchWithCache performs search with Redis caching
func (e *GoogleEngine) SearchWithCache(ctx context.Context, rdb *redis.Client, query string, limit int, ttl time.Duration) ([]SearchResult, error) {
	cacheKey := fmt.Sprintf("dork:google:%s", query)

	// Try cache first
	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Printf("Cache hit for query: %s", query)
			// Note: In production, use proper JSON unmarshaling
		}
		_ = cached // For now, always do fresh search
	}

	results, err := e.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// Cache results
	if rdb != nil && len(results) > 0 {
		// Note: In production, marshal results to JSON
		rdb.Set(ctx, cacheKey, fmt.Sprintf("%d results", len(results)), ttl)
	}

	return results, nil
}

// GetEngineName returns the engine name
func (e *GoogleEngine) GetEngineName() string {
	return "google"
}
