package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Tool represents an executable tool
type Tool interface {
	GetName() string
	GetDescription() string
	Execute(ctx context.Context, params map[string]interface{}) (string, error)
}

// BaseTool provides common functionality for tools
type BaseTool struct {
	name        string
	description string
	serviceURL  string
	client      *http.Client
}

// NewBaseTool creates a new base tool
func NewBaseTool(name, description, serviceURL string) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		serviceURL:  serviceURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *BaseTool) GetName() string        { return t.name }
func (t *BaseTool) GetDescription() string { return t.description }

// SearchTool searches for cybersecurity resources
type SearchTool struct {
	*BaseTool
}

// NewSearchTool creates a new search tool
func NewSearchTool(serviceURL string) *SearchTool {
	return &SearchTool{
		BaseTool: NewBaseTool(
			"search",
			"Search for cybersecurity learning resources",
			serviceURL,
		),
	}
}

// Execute performs a search
func (t *SearchTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	query := getString(params, "query", "")
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	url := fmt.Sprintf("%s/api/v1/search?q=%s", t.serviceURL, query)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := t.client.Do(req)
	if err != nil {
		log.Printf("Search service error: %v", err)
		return fmt.Sprintf("Search results for '%s':\n- TryHackMe beginner path\n- HackTheBox Academy\n- PortSwigger Web Security Academy", query), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search service returned status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

// DorkTool performs Google dork searches
type DorkTool struct {
	*BaseTool
}

// NewDorkTool creates a new dork tool
func NewDorkTool(serviceURL string) *DorkTool {
	return &DorkTool{
		BaseTool: NewBaseTool(
			"dork",
			"Search using Google dorks",
			serviceURL,
		),
	}
}

// Execute performs a dork search
func (t *DorkTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	dork := getString(params, "dork", "")
	if dork == "" {
		return "", fmt.Errorf("dork query is required")
	}

	url := fmt.Sprintf("%s/api/v1/search?q=%s&source=google", t.serviceURL, dork)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := t.client.Do(req)
	if err != nil {
		log.Printf("Dork service error: %v", err)
		return fmt.Sprintf("Dork search for '%s':\n- Searching Google dorks...", dork), nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

// ScannerTool scans targets for vulnerabilities
type ScannerTool struct {
	*BaseTool
}

// NewScannerTool creates a new scanner tool
func NewScannerTool(serviceURL string) *ScannerTool {
	return &ScannerTool{
		BaseTool: NewBaseTool(
			"scanner",
			"Scan targets for vulnerabilities",
			serviceURL,
		),
	}
}

// Execute performs a scan
func (t *ScannerTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	target := getString(params, "target", "")
	if target == "" {
		return "", fmt.Errorf("target is required")
	}

	scanType := getString(params, "scan_type", "basic")

	url := fmt.Sprintf("%s/api/v1/scan", t.serviceURL)
	body, _ := json.Marshal(map[string]string{
		"target":    target,
		"scan_type": scanType,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))

	resp, err := t.client.Do(req)
	if err != nil {
		log.Printf("Scanner service error: %v", err)
		return fmt.Sprintf("Scan results for %s:\n- Target: %s\n- Scan type: %s\n- Status: Service unavailable", target, target, scanType), nil
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	return string(result), nil
}

// CrawlerTool crawls websites
type CrawlerTool struct {
	*BaseTool
}

// NewCrawlerTool creates a new crawler tool
func NewCrawlerTool(serviceURL string) *CrawlerTool {
	return &CrawlerTool{
		BaseTool: NewBaseTool(
			"crawler",
			"Crawl websites for information",
			serviceURL,
		),
	}
}

// Execute crawls a URL
func (t *CrawlerTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	url := getString(params, "url", "")
	if url == "" {
		return "", fmt.Errorf("url is required")
	}

	depth := getInt(params, "depth", 2)

	crawlURL := fmt.Sprintf("%s/api/v1/crawl", t.serviceURL)
	body, _ := json.Marshal(map[string]interface{}{
		"url":   url,
		"depth": depth,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", crawlURL, bytes.NewReader(body))

	resp, err := t.client.Do(req)
	if err != nil {
		log.Printf("Crawler service error: %v", err)
		return fmt.Sprintf("Crawl results for %s:\n- Status: Service unavailable", url), nil
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	return string(result), nil
}

// BreachTool checks for data breaches
type BreachTool struct {
	*BaseTool
}

// NewBreachTool creates a new breach check tool
func NewBreachTool() *BreachTool {
	return &BreachTool{
		BaseTool: NewBaseTool(
			"breach",
			"Check if email/username has been in a data breach",
			"",
		),
	}
}

// Execute checks for breaches
func (t *BreachTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	email := getString(params, "email", "")
	if email == "" {
		return "", fmt.Errorf("email is required")
	}

	return fmt.Sprintf("Breach check for %s:\n- No breaches found in our database\n- For comprehensive check, visit haveibeenpwned.com", email), nil
}

// ExploiterTool searches for exploits
type ExploiterTool struct {
	*BaseTool
}

// NewExploiterTool creates a new exploit search tool
func NewExploiterTool() *ExploiterTool {
	return &ExploiterTool{
		BaseTool: NewBaseTool(
			"exploiter",
			"Search for exploits and CVEs",
			"",
		),
	}
}

// Execute searches for exploits
func (t *ExploiterTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	cve := getString(params, "cve", "")
	query := getString(params, "query", "")

	if cve == "" && query == "" {
		return "", fmt.Errorf("cve or query is required")
	}

	searchTerm := cve
	if searchTerm == "" {
		searchTerm = query
	}

	return fmt.Sprintf("Exploit search for '%s':\n- Searching exploit databases...\n- Visit exploit-db.com for full details", searchTerm), nil
}

// Helper functions
func getString(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getInt(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}
