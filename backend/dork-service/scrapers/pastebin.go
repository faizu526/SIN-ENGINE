package scrapers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PastebinScraper scrapes Pastebin and similar sites
type PastebinScraper struct {
	Scraper
	baseURL string
	sites   []string
}

// NewPastebinScraper creates a Pastebin scraper
func NewPastebinScraper() *PastebinScraper {
	p := &PastebinScraper{
		baseURL: "https://pastebin.com",
		sites: []string{
			"https://pastebin.com",
			"https://gist.github.com",
			"https://paste.ee",
			"https://hastebin.com",
		},
	}
	p.client = &http.Client{
		Timeout: 30 * time.Second,
	}
	return p
}

// Search searches paste sites for leaks
func (p *PastebinScraper) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	results := make([]SearchResult, 0)

	// Search Pastebin
	pastebinResults, err := p.searchPastebin(ctx, query, limit)
	if err != nil {
		log.Printf("Pastebin search error: %v", err)
	} else {
		results = append(results, pastebinResults...)
	}

	// Search GitHub Gist
	githubResults, err := p.searchGitHubGist(ctx, query, limit)
	if err != nil {
		log.Printf("GitHub Gist search error: %v", err)
	} else {
		results = append(results, githubResults...)
	}

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	log.Printf("Pastebin search for '%s' returned %d results", query, len(results))
	return results, nil
}

// searchPastebin searches Pastebin
func (p *PastebinScraper) searchPastebin(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	results := make([]SearchResult, 0)

	// Use Pastebin's raw archive search
	searchURL := fmt.Sprintf("%s/archive?sort=recent&q=%s", p.baseURL, url.QueryEscape(query))

	doc, err := p.fetch(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	doc.Find(".post-link").Each(func(i int, s *goquery.Selection) {
		if len(results) >= limit {
			return
		}

		link, _ := s.Attr("href")
		title := s.Text()

		if link == "" || title == "" {
			return
		}

		results = append(results, SearchResult{
			Title:   title,
			URL:     "https://pastebin.com" + link,
			Source:  "pastebin",
			Snippet: "Recent paste matching: " + query,
		})
	})

	return results, nil
}

// searchGitHubGist searches GitHub Gist
func (p *PastebinScraper) searchGitHubGist(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	results := make([]SearchResult, 0)

	// Use GitHub's search API for gists
	apiURL := fmt.Sprintf("https://api.github.com/search/code?q=%s+in:file+extension:gist", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "SIN-Engine")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to web scraping
		return p.searchGitHubGistWeb(ctx, query, limit)
	}

	// Parse JSON response
	var gistResp struct {
		Items []struct {
			HTMLURL string `json:"html_url"`
			Path    string `json:"path"`
		} `json:"items"`
	}

	// Simple parsing - just return empty if can't parse
	_ = gistResp

	// Fallback: scrape GitHub search results
	return p.searchGitHubGistWeb(ctx, query, limit)
}

// searchGitHubGistWeb scrapes GitHub Gist search
func (p *PastebinScraper) searchGitHubGistWeb(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	results := make([]SearchResult, 0)

	searchURL := fmt.Sprintf("https://gist.github.com/search?q=%s", url.QueryEscape(query))

	doc, err := p.fetch(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	doc.Find(".gist-snippet").Each(func(i int, s *goquery.Selection) {
		if len(results) >= limit {
			return
		}

		title := s.Find(".gist-snippet-meta li").First().Text()
		linkElem := s.Find("a").First()
		link, _ := linkElem.Attr("href")

		if link == "" {
			return
		}

		results = append(results, SearchResult{
			Title:   strings.TrimSpace(title),
			URL:     "https://gist.github.com" + link,
			Source:  "github-gist",
			Snippet: "Gist containing: " + query,
		})
	})

	return results, nil
}

// PastebinHandler handles paste search requests
type PastebinHandler struct {
	scraper *PastebinScraper
}

// NewPastebinHandler creates a new handler
func NewPastebinHandler() *PastebinHandler {
	return &PastebinHandler{
		scraper: NewPastebinScraper(),
	}
}

// Search performs the search
func (h *PastebinHandler) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return h.scraper.Search(ctx, query, limit)
}

// SearchBySite searches a specific paste site
func (h *PastebinHandler) SearchBySite(ctx context.Context, site string, query string, limit int) ([]SearchResult, error) {
	switch site {
	case "pastebin":
		return h.scraper.searchPastebin(ctx, query, limit)
	case "gist":
		return h.scraper.searchGitHubGist(ctx, query, limit)
	default:
		return h.scraper.Search(ctx, query, limit)
	}
}
