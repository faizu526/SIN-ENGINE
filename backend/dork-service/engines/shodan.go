package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
)

// ShodanEngine handles Shodan API searches
type ShodanEngine struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewShodanEngine creates a new Shodan search engine
func NewShodanEngine(apiKey string) *ShodanEngine {
	return &ShodanEngine{
		apiKey:  apiKey,
		baseURL: "https://api.shodan.io",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ShodanHost represents a host result from Shodan
type ShodanHost struct {
	IP        string   `json:"ip_str"`
	Port      int      `json:"port"`
	Protocol  string   `json:"protocol"`
	Transport string   `json:"transport"`
	Product   string   `json:"product"`
	Version   string   `json:"version"`
	Hostnames []string `json:"hostnames"`
	Org       string   `json:"org"`
	OS        string   `json:"os"`
	Timestamp string   `json:"timestamp"`
	Data      string   `json:"data"`
	Country   string   `json:"country_name"`
	City      string   `json:"city"`
	ISP       string   `json:"isp"`
	ASN       string   `json:"asn"`
	_vulns    []string `json:"vulns"`
}

// ShodanSearchResponse represents the API response
type ShodanSearchResponse struct {
	Total   int          `json:"total"`
	Page    int          `json:"page"`
	Results []ShodanHost `json:"matches"`
}

// Search performs a Shodan search
func (e *ShodanEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("Shodan API key not configured")
	}

	if limit <= 0 {
		limit = 10
	}

	// Build API URL
	apiURL := fmt.Sprintf("%s/shodan/host/search?key=%s&query=%s&limit=%d",
		e.baseURL, e.apiKey, url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Shodan API error: %s", string(body))
	}

	var shodanResp ShodanSearchResponse
	if err := json.Unmarshal(body, &shodanResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(shodanResp.Results))
	for _, host := range shodanResp.Results {
		title := fmt.Sprintf("%s:%d - %s", host.IP, host.Port, host.Product)
		if title == ":" {
			title = fmt.Sprintf("%s:%d", host.IP, host.Port)
		}

		snippet := host.Data
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}

		results = append(results, SearchResult{
			Title:     title,
			URL:       fmt.Sprintf("https://www.shodan.io/host/%s", host.IP),
			Snippet:   snippet,
			Engine:    "shodan",
			Category:  host.Product,
			IndexedAt: time.Now().Unix(),
		})
	}

	log.Printf("Shodan search for '%s' returned %d results", query, len(results))
	return results, nil
}

// SearchHost performs a detailed host search
func (e *ShodanEngine) SearchHost(ctx context.Context, ip string) (*ShodanHost, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("Shodan API key not configured")
	}

	apiURL := fmt.Sprintf("%s/shodan/host/%s?key=%s", e.baseURL, ip, e.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Shodan API error: status %d", resp.StatusCode)
	}

	var host ShodanHost
	if err := json.NewDecoder(resp.Body).Decode(&host); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &host, nil
}

// SearchWithCache performs search with Redis caching
func (e *ShodanEngine) SearchWithCache(ctx context.Context, rdb *redis.Client, query string, limit int, ttl time.Duration) ([]SearchResult, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("Shodan API key not configured")
	}

	cacheKey := fmt.Sprintf("dork:shodan:%s", query)

	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Printf("Cache hit for Shodan query: %s", query)
			// Note: In production, unmarshal cached JSON
			_ = cached
		}
	}

	results, err := e.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	if rdb != nil && len(results) > 0 {
		rdb.Set(ctx, cacheKey, fmt.Sprintf("%d results", len(results)), ttl)
	}

	return results, nil
}

// GetEngineName returns the engine name
func (e *ShodanEngine) GetEngineName() string {
	return "shodan"
}
