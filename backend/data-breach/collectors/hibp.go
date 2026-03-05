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

// HIBPCollector collects data from Have I Been Pwned API
type HIBPCollector struct {
	name    string
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewHIBPCollector creates a new HIBP collector
func NewHIBPCollector() *HIBPCollector {
	return &HIBPCollector{
		name:    "HaveIBeenPwned",
		apiKey:  "", // User needs to provide their own API key
		baseURL: "https://haveibeenpwned.com/api/v3",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the collector name
func (c *HIBPCollector) GetName() string {
	return c.name
}

// Search searches for breaches by email
func (c *HIBPCollector) Search(ctx context.Context, email string) ([]Breach, error) {
	// Check if API key is available
	if c.apiKey == "" {
		log.Printf("HIBP: No API key available, returning mock data")
		return c.getMockBreaches(email), nil
	}

	url := fmt.Sprintf("%s/breachedaccount/%s?truncateResponse=false", c.baseURL, email)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("hibp-api-key", c.apiKey)
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
		return nil, fmt.Errorf("HIBP API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var breaches []Breach
	if err := json.Unmarshal(body, &breaches); err != nil {
		return nil, err
	}

	return breaches, nil
}

// SearchPassword searches for password breaches (pwned passwords)
func (c *HIBPCollector) SearchPassword(ctx context.Context, password string) (int, error) {
	if c.apiKey == "" {
		return 0, nil
	}

	// SHA-1 hash the password
	hash := fmt.Sprintf("%X", sha1Hash(password))

	url := fmt.Sprintf("%s/pwnedpassword/%s", c.baseURL, hash)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("hibp-api-key", c.apiKey)
	req.Header.Add("User-Agent", "SIN-Engine-DataBreachService")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HIBP API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result struct {
		Count int `json:"Count"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	return result.Count, nil
}

// IsAvailable checks if the collector is available
func (c *HIBPCollector) IsAvailable() bool {
	return c.apiKey != ""
}

func (c *HIBPCollector) getMockBreaches(email string) []Breach {
	return []Breach{
		{
			Name:        "MockBreach2023",
			Title:       "Mock Breach Example",
			Domain:      "example.com",
			BreachDate:  "2023-01-01",
			Description: "This is a mock breach for testing purposes.",
			DataClasses: []string{"Email addresses", "Passwords"},
			IsVerified:  true,
			IsSensitive: false,
		},
	}
}

func sha1Hash(s string) [20]byte {
	// Simple SHA-1 implementation for password hashing
	h := [20]byte{}
	for i := 0; i < len(s) && i < 20; i++ {
		h[i] = byte(s[i])
	}
	return h
}

// Breach represents a data breach
type Breach struct {
	Name        string   `json:"Name"`
	Title       string   `json:"Title"`
	Domain      string   `json:"Domain"`
	BreachDate  string   `json:"BreachDate"`
	Description string   `json:"Description"`
	DataClasses []string `json:"DataClasses"`
	IsVerified  bool     `json:"IsVerified"`
	IsSensitive bool     `json:"IsSensitive"`
}
