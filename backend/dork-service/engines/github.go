package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// GitHubEngine handles GitHub code search
type GitHubEngine struct {
	token   string
	baseURL string
	client  *http.Client
}

// NewGitHubEngine creates a new GitHub search engine
func NewGitHubEngine() *GitHubEngine {
	return &GitHubEngine{
		token:   "",
		baseURL: "https://api.github.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	URL         string `json:"html_url"`
	Stars       int    `json:"stargazers_count"`
	Language    string `json:"language"`
	UpdatedAt   string `json:"updated_at"`
}

// GitHubCodeResult represents a code search result
type GitHubCodeResult struct {
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	URL        string  `json:"html_url"`
	Repository string  `json:"repository"`
	Language   string  `json:"language"`
	Score      float64 `json:"score"`
}

// SearchRepos searches GitHub repositories
func (e *GitHubEngine) SearchRepos(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	apiURL := fmt.Sprintf("%s/search/repositories?q=%s&per_page=%d&sort=stars&order=desc",
		e.baseURL, url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if e.token != "" {
		req.Header.Set("Authorization", "token "+e.token)
	}
	req.Header.Set("User-Agent", "SIN-Engine-Dork-Service")

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
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var githubResp map[string]interface{}
	if err := json.Unmarshal(body, &githubResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, limit)

	items, ok := githubResp["items"].([]interface{})
	if !ok {
		log.Printf("GitHub search for '%s' - no results", query)
		return results, nil
	}

	for _, item := range items {
		if len(results) >= limit {
			break
		}

		repo, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := repo["full_name"].(string)
		desc, _ := repo["description"].(string)
		repoURL, _ := repo["html_url"].(string)
		stars, _ := repo["stargazers_count"].(float64)
		lang, _ := repo["language"].(string)

		snippet := desc
		if snippet == "" {
			snippet = fmt.Sprintf("⭐ %.0f stars | %s", stars, lang)
		} else {
			snippet = fmt.Sprintf("%s | ⭐ %.0f stars | %s", desc, stars, lang)
		}

		results = append(results, SearchResult{
			Title:     name,
			URL:       repoURL,
			Snippet:   snippet,
			Engine:    "github",
			Category:  lang,
			IndexedAt: time.Now().Unix(),
		})
	}

	log.Printf("GitHub repo search for '%s' returned %d results", query, len(results))
	return results, nil
}

// SearchCode searches GitHub code
func (e *GitHubEngine) SearchCode(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Build code search query
	codeQuery := url.QueryEscape(query + " in:file")

	apiURL := fmt.Sprintf("%s/search/code?q=%s&per_page=%d",
		e.baseURL, codeQuery, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if e.token != "" {
		req.Header.Set("Authorization", "token "+e.token)
	}
	req.Header.Set("User-Agent", "SIN-Engine-Dork-Service")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for rate limiting
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API rate limited")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var codeResp map[string]interface{}
	if err := json.Unmarshal(body, &codeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, limit)

	items, ok := codeResp["items"].([]interface{})
	if !ok {
		log.Printf("GitHub code search for '%s' - no results", query)
		return results, nil
	}

	for _, item := range items {
		if len(results) >= limit {
			break
		}

		code, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := code["name"].(string)
		path, _ := code["path"].(string)
		repoURL, _ := code["repository"].(map[string]interface{})
		repoName, _ := repoURL["full_name"].(string)
		fileURL, _ := code["html_url"].(string)
		lang, _ := code["language"].(string)

		title := fmt.Sprintf("%s in %s", name, repoName)
		snippet := fmt.Sprintf("Path: %s | Language: %s", path, lang)

		results = append(results, SearchResult{
			Title:     title,
			URL:       fileURL,
			Snippet:   snippet,
			Engine:    "github",
			Category:  lang,
			IndexedAt: time.Now().Unix(),
		})
	}

	log.Printf("GitHub code search for '%s' returned %d results", query, len(results))
	return results, nil
}

// Search performs a GitHub search (repos by default)
func (e *GitHubEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Check if it's a code search
	if strings.Contains(query, "in:file") || strings.Contains(query, "extension:") {
		return e.SearchCode(ctx, query, limit)
	}
	return e.SearchRepos(ctx, query, limit)
}

// SearchWithCache performs search with Redis caching
func (e *GitHubEngine) SearchWithCache(ctx context.Context, rdb *redis.Client, query string, limit int, ttl time.Duration) ([]SearchResult, error) {
	cacheKey := fmt.Sprintf("dork:github:%s", query)

	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Printf("Cache hit for GitHub query: %s", query)
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
func (e *GitHubEngine) GetEngineName() string {
	return "github"
}
