package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// SnusbaseCollector collects data from Snusbase API
type SnusbaseCollector struct {
	name    string
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewSnusbaseCollector creates a new Snusbase collector
func NewSnusbaseCollector() *SnusbaseCollector {
	return &SnusbaseCollector{
		name:    "Snusbase",
		apiKey:  "", // User needs to provide their own API key
		baseURL: "https://api.snusbase.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the collector name
func (c *SnusbaseCollector) GetName() string {
	return c.name
}

// Search searches for breaches
func (c *SnusbaseCollector) Search(ctx context.Context, query string) ([]Breach, error) {
	if c.apiKey == "" {
		log.Printf("Snusbase: No API key available, returning mock data")
		return c.getMockBreaches(query), nil
	}

	url := fmt.Sprintf("%s/v3/search?term=%s&type=email", c.baseURL, query)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Auth", c.apiKey)
	req.Header.Add("User-Agent", "SIN-Engine-DataBreachService")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []Breach{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Snusbase API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []Breach `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// IsAvailable checks if the collector is available
func (c *SnusbaseCollector) IsAvailable() bool {
	return c.apiKey != ""
}

func (c *SnusbaseCollector) getMockBreaches(query string) []Breach {
	return []Breach{
		{
			Name:        "SnusbaseMock",
			Title:       "Snusbase Mock Breach",
			Domain:      "snusbase.mock",
			BreachDate:  "2023-09-10",
			Description: "Mock breach data from Snusbase collector",
			DataClasses: []string{"Email", "Username", "Password"},
			IsVerified:  true,
			IsSensitive: false,
		},
	}
}
