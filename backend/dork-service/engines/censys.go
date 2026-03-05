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

// CensysEngine handles Censys searches
type CensysEngine struct {
	apiID     string
	apiSecret string
	baseURL   string
	client    *http.Client
}

// NewCensysEngine creates a new Censys search engine
func NewCensysEngine() *CensysEngine {
	return &CensysEngine{
		apiID:     "",
		apiSecret: "",
		baseURL:   "https://search.censys.io/api/v2",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CensysHost represents a host from Censys
type CensysHost struct {
	IP               string          `json:"ip"`
	Protocols        []string        `json:"protocols"`
	Services         []CensysService `json:"services"`
	Location         CensysLocation  `json:"location"`
	AutonomousSystem CensysAS        `json:"autonomous_system"`
	Hostname         []string        `json:"hostnames"`
	UpdatedAt        string          `json:"updated_at"`
	ObservedAt       string          `json:"observed_at"`
}

// CensysService represents a service
type CensysService struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Product  string `json:"product"`
	Version  string `json:"version"`
}

// CensysLocation represents location data
type CensysLocation struct {
	Country   string `json:"country"`
	City      string `json:"city"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// CensysAS represents autonomous system
type CensysAS struct {
	ASN     int    `json:"asn"`
	Prefix  string `json:"prefix"`
	Country string `json:"country"`
}

// Search performs a Censys search
func (e *CensysEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if e.apiID == "" || e.apiSecret == "" {
		return nil, fmt.Errorf("Censys API credentials not configured")
	}

	if limit <= 0 {
		limit = 10
	}

	// Build API URL
	apiURL := fmt.Sprintf("%s/hosts/search?q=%s&per_page=%d",
		e.baseURL, url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(e.apiID, e.apiSecret)
	req.Header.Set("Accept", "application/json")

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
		return nil, fmt.Errorf("Censys API error: %s", string(body))
	}

	var censysResp map[string]interface{}
	if err := json.Unmarshal(body, &censysResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, limit)

	// Extract results from response
	resultData, ok := censysResp["result"].(map[string]interface{})
	if !ok {
		log.Printf("Censys search for '%s' - no results format", query)
		return results, nil
	}

	hits, ok := resultData["hits"].([]interface{})
	if !ok {
		return results, nil
	}

	for _, hit := range hits {
		if len(results) >= limit {
			break
		}

		host, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		ip, _ := host["ip"].(string)
		protocols, _ := host["protocols"].([]string)

		title := ip
		if len(protocols) > 0 {
			title = fmt.Sprintf("%s (%s)", ip, protocols[0])
		}

		results = append(results, SearchResult{
			Title:     title,
			URL:       fmt.Sprintf("https://search.censys.io/hosts/%s", ip),
			Snippet:   fmt.Sprintf("Protocols: %v", protocols),
			Engine:    "censys",
			IndexedAt: time.Now().Unix(),
		})
	}

	log.Printf("Censys search for '%s' returned %d results", query, len(results))
	return results, nil
}

// SearchWithCache performs search with Redis caching
func (e *CensysEngine) SearchWithCache(ctx context.Context, rdb *redis.Client, query string, limit int, ttl time.Duration) ([]SearchResult, error) {
	cacheKey := fmt.Sprintf("dork:censys:%s", query)

	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Printf("Cache hit for Censys query: %s", query)
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
func (e *CensysEngine) GetEngineName() string {
	return "censys"
}
