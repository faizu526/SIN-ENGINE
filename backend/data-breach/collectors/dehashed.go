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

// DehashedCollector collects data from Dehashed API
type DehashedCollector struct {
	name    string
	apiKey  string
	email   string
	baseURL string
	client  *http.Client
}

// NewDehashedCollector creates a new Dehashed collector
func NewDehashedCollector() *DehashedCollector {
	return &DehashedCollector{
		name:    "Dehashed",
		apiKey:  "", // User needs to provide their own API key
		email:   "",
		baseURL: "https://api.dehashed.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the collector name
func (c *DehashedCollector) GetName() string {
	return c.name
}

// Search searches for breaches by email
func (c *DehashedCollector) Search(ctx context.Context, query string) ([]Breach, error) {
	if c.apiKey == "" {
		log.Printf("Dehashed: No API key available, returning mock data")
		return c.getMockBreaches(query), nil
	}

	url := fmt.Sprintf("%s/search?query=%s", c.baseURL, query)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.email, c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []Breach{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Dehashed API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Entries []Breach `json:"entries"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Entries, nil
}

// IsAvailable checks if the collector is available
func (c *DehashedCollector) IsAvailable() bool {
	return c.apiKey != ""
}

func (c *DehashedCollector) getMockBreaches(query string) []Breach {
	return []Breach{
		{
			Name:        "DehashedMock",
			Title:       "Dehashed Mock Breach",
			Domain:      "mockdomain.com",
			BreachDate:  "2023-06-15",
			Description: "Mock breach data from Dehashed collector",
			DataClasses: []string{"Email", "Password", "Phone"},
			IsVerified:  true,
			IsSensitive: false,
		},
	}
}
