package scrapers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Scraper is a base struct for web scrapers
type Scraper struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

// NewScraper creates a new base scraper
func NewScraper(timeout time.Duration) *Scraper {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Scraper{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		timeout: timeout,
	}
}

// fetch makes an HTTP GET request and returns a goquery document
func (s *Scraper) fetch(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}

// SearchResult represents a scraped result
type SearchResult struct {
	Title   string
	URL     string
	Source  string
	Snippet string
	Date    string
}

// Search queries a scraper
func (s *Scraper) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return nil, fmt.Errorf("Search not implemented")
}

// SetBaseURL sets the base URL for the scraper
func (s *Scraper) SetBaseURL(url string) {
	s.baseURL = url
}

// RateLimiter implements simple rate limiting
type RateLimiter struct {
	requests chan struct{}
}

// NewRateLimiter creates a rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 1
	}

	rl := &RateLimiter{
		requests: make(chan struct{}, requestsPerSecond),
	}

	// Start rate limiting goroutine
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))
		for range ticker.C {
			select {
			case rl.requests <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

// Acquire acquires a rate limit token
func (rl *RateLimiter) Acquire() {
	<-rl.requests
}

// Release releases the token (no-op for token bucket)
func (rl *RateLimiter) Release() {}

// ProxyScraper wraps a scraper with proxy support
type ProxyScraper struct {
	Scraper
	proxies []string
	index   int
}

// NewProxyScraper creates a scraper with proxy rotation
func NewProxyScraper(proxies []string) *ProxyScraper {
	ps := &ProxyScraper{
		proxies: proxies,
		index:   0,
	}
	ps.client = &http.Client{
		Timeout: 30 * time.Second,
	}
	return ps
}

// GetNextProxy returns the next proxy in rotation
func (ps *ProxyScraper) GetNextProxy() string {
	if len(ps.proxies) == 0 {
		return ""
	}
	proxy := ps.proxies[ps.index]
	ps.index = (ps.index + 1) % len(ps.proxies)
	return proxy
}

// Cache stores scraped data
type Cache struct {
	data map[string]cacheEntry
}

type cacheEntry struct {
	result  []SearchResult
	expires time.Time
}

// NewCache creates a new cache
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		data: make(map[string]cacheEntry),
	}

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(ttl)
		for range ticker.C {
			c.cleanup()
		}
	}()

	return c
}

// Get retrieves cached results
func (c *Cache) Get(key string) ([]SearchResult, bool) {
	entry, ok := c.data[key]
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}
	return entry.result, true
}

// Set stores results in cache
func (c *Cache) Set(key string, result []SearchResult, ttl time.Duration) {
	c.data[key] = cacheEntry{
		result:  result,
		expires: time.Now().Add(ttl),
	}
}

func (c *Cache) cleanup() {
	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.expires) {
			delete(c.data, key)
		}
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
