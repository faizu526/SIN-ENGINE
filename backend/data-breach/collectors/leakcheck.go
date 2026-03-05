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

// LeakCheckCollector collects data from LeakCheck API
type LeakCheckCollector struct {
	name    string
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewLeakCheckCollector creates a new LeakCheck collector
func NewLeakCheckCollector() *LeakCheckCollector {
	return &LeakCheckCollector{
		name:    "LeakCheck",
		apiKey:  "", // User needs to provide their own API key
		baseURL: "https://leakcheck.io/api",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the collector name
func (c *LeakCheckCollector) GetName() string {
	return c.name
}

// Search searches for breaches
func (c *LeakCheckCollector) Search(ctx context.Context, query string) ([]Breach, error) {
	if c.apiKey == "" {
		log.Printf("LeakCheck: No API key available, returning mock data")
		return c.getMockBreaches(query), nil
	}

	url := fmt.Sprintf("%s/public/%s?key=%s", c.baseURL, query, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []Breach{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LeakCheck API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Breaches []Breach `json:"sources"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Breaches, nil
}

// IsAvailable checks if the collector is available
func (c *LeakCheckCollector) IsAvailable() bool {
	return c.apiKey != ""
}

func (c *LeakCheckCollector) getMockBreaches(query string) []Breach {
	return []Breach{
		{
			Name:        "LeakCheckMock",
			Title:       "LeakCheck Mock Breach",
			Domain:      "leakcheck.mock",
			BreachDate:  "2023-08-20",
			Description: "Mock breach data from LeakCheck collector",
			DataClasses: []string{"Email", "Password", "IP"},
			IsVerified:  true,
			IsSensitive: false,
		},
	}
}
