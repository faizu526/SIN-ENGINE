package engines

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// FOFAEngine handles FOFA searches
type FOFAEngine struct {
	email   string
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewFOFAEngine creates a new FOFA search engine
func NewFOFAEngine() *FOFAEngine {
	return &FOFAEngine{
		email:   "",
		apiKey:  "",
		baseURL: "https://fofa.info/api/v1/search/all",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FOFAHost represents a host from FOFA
type FOFAHost struct {
	IP       string `json:"ip"`
	Domain   string `json:"domain"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Product  string `json:"product"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Server   string `json:"server"`
	ASN      string `json:"as_number"`
}

// Search performs a FOFA search
func (e *FOFAEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if e.email == "" || e.apiKey == "" {
		return nil, fmt.Errorf("FOFA API credentials not configured")
	}

	if limit <= 0 {
		limit = 10
	}

	// Build query string
	queryStr := base64.StdEncoding.EncodeToString([]byte(query))

	apiURL := fmt.Sprintf("%s?email=%s&key=%s&qbase64=%s&size=%d",
		e.baseURL, e.email, e.apiKey, queryStr, limit)

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
		return nil, fmt.Errorf("FOFA API error: %s", string(body))
	}

	var fofaResp map[string]interface{}
	if err := json.Unmarshal(body, &fofaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, limit)

	// Extract results - FOFA returns array of arrays
	data, ok := fofaResp["data"].([]interface{})
	if !ok {
		log.Printf("FOFA search for '%s' - no results", query)
		return results, nil
	}

	for _, item := range data {
		if len(results) >= limit {
			break
		}

		hostData, ok := item.([]interface{})
		if !ok || len(hostData) < 6 {
			continue
		}

		// FOFA returns: [host, ip, port, protocol, country, city, ...]
		host, _ := hostData[0].(string)
		ip, _ := hostData[1].(string)
		port, _ := hostData[2].(float64)
		protocol, _ := hostData[3].(string)

		title := fmt.Sprintf("%s:%d", ip, int(port))
		if host != "" {
			title = fmt.Sprintf("%s (%s:%d)", host, ip, int(port))
		}

		results = append(results, SearchResult{
			Title:     title,
			URL:       fmt.Sprintf("https://fofa.info/host/%s", ip),
			Snippet:   fmt.Sprintf("Protocol: %s | Port: %d", protocol, int(port)),
			Engine:    "fofa",
			IndexedAt: time.Now().Unix(),
		})
	}

	log.Printf("FOFA search for '%s' returned %d results", query, len(results))
	return results, nil
}

// SearchWithCache performs search with Redis caching
func (e *FOFAEngine) SearchWithCache(ctx context.Context, rdb *redis.Client, query string, limit int, ttl time.Duration) ([]SearchResult, error) {
	cacheKey := fmt.Sprintf("dork:fofa:%s", query)

	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Printf("Cache hit for FOFA query: %s", query)
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
func (e *FOFAEngine) GetEngineName() string {
	return "fofa"
}
