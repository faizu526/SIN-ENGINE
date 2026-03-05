package engine

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CrawlerEngine struct {
	db          *gorm.DB
	redisClient *redis.Client
	httpClient  *http.Client
	jobs        map[string]*CrawlJob
	mu          sync.RWMutex
	active      bool
}

type CrawlJob struct {
	ID         string
	URL        string
	Depth      int
	MaxDepth   int
	UserAgent  string
	Concurrent int
	Timeout    int
	Filters    []string
	Status     string
	Results    []CrawlResult
	StartTime  time.Time
	EndTime    time.Time
	Error      error
}

type CrawlResult struct {
	URL        string            `json:"url"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Links      []string          `json:"links"`
	Images     []string          `json:"images"`
	Meta       map[string]string `json:"meta"`
	Depth      int               `json:"depth"`
	CrawledAt  time.Time         `json:"crawled_at"`
	StatusCode int               `json:"status_code"`
}

func NewCrawlerEngine(db *gorm.DB, redisClient *redis.Client) *CrawlerEngine {
	return &CrawlerEngine{
		db:          db,
		redisClient: redisClient,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		jobs:   make(map[string]*CrawlJob),
		active: true,
	}
}

func (e *CrawlerEngine) StartJob(ctx context.Context, job *CrawlJob) error {
	job.Status = "running"
	job.StartTime = time.Now()

	e.mu.Lock()
	e.jobs[job.ID] = job
	e.mu.Unlock()

	// Save job to Redis for persistence
	if e.redisClient != nil {
		e.redisClient.Set(ctx, "crawl:job:"+job.ID, "running", 24*time.Hour)
	}

	log.Printf("Starting crawl job %s for URL: %s", job.ID, job.URL)

	go func() {
		e.crawl(ctx, job)
	}()

	return nil
}

func (e *CrawlerEngine) crawl(ctx context.Context, job *CrawlJob) {
	defer func() {
		job.EndTime = time.Now()
		job.Status = "completed"

		e.mu.Lock()
		delete(e.jobs, job.ID)
		e.mu.Unlock()

		if e.redisClient != nil {
			e.redisClient.Set(ctx, "crawl:job:"+job.ID, "completed", 24*time.Hour)
		}

		log.Printf("Crawl job %s completed. Found %d results in %v",
			job.ID, len(job.Results), job.EndTime.Sub(job.StartTime))
	}()

	e.crawlURL(ctx, job, job.URL, 0)
}

func (e *CrawlerEngine) crawlURL(ctx context.Context, job *CrawlJob, url string, depth int) {
	if depth > job.MaxDepth {
		return
	}

	// Check if URL passes filters
	if !e.passesFilters(url, job.Filters) {
		return
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for %s: %v", url, err)
		return
	}

	if job.UserAgent != "" {
		req.Header.Set("User-Agent", job.UserAgent)
	} else {
		req.Header.Set("User-Agent", "SIN-Engine-Crawler/1.0")
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		log.Printf("Error crawling %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Error parsing HTML from %s: %v", url, err)
		return
	}

	// Extract data
	result := CrawlResult{
		URL:        url,
		Depth:      depth,
		CrawledAt:  time.Now(),
		StatusCode: resp.StatusCode,
	}

	// Get title
	result.Title = doc.Find("title").First().Text()

	// Get meta description
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		prop, _ := s.Attr("property")
		if name == "description" || prop == "og:description" {
			content, _ := s.Attr("content")
			if result.Meta == nil {
				result.Meta = make(map[string]string)
			}
			result.Meta["description"] = content
		}
	})

	// Get links
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && href != "" && !strings.HasPrefix(href, "#") && !strings.HasPrefix(href, "javascript:") {
			result.Links = append(result.Links, href)
		}
	})

	// Get images
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && src != "" {
			result.Images = append(result.Images, src)
		}
	})

	// Get text content (limited)
	var text strings.Builder
	doc.Find("p, h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		text.WriteString(s.Text())
		text.WriteString(" ")
	})
	result.Content = text.String()

	// Add result
	job.Results = append(job.Results, result)

	// Save to Redis
	if e.redisClient != nil {
		e.redisClient.RPush(ctx, "crawl:results:"+job.ID, url)
	}

	// Continue crawling links if within depth limit
	if depth < job.MaxDepth {
		for _, link := range result.Links {
			select {
			case <-ctx.Done():
				return
			default:
				// Convert relative URLs to absolute
				absoluteURL := resolveURL(job.URL, link)
				if absoluteURL != "" {
					e.crawlURL(ctx, job, absoluteURL, depth+1)
				}
			}
		}
	}
}

func (e *CrawlerEngine) passesFilters(url string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		if strings.Contains(url, filter) {
			return true
		}
	}
	return false
}

func resolveURL(base, relative string) string {
	// If already absolute
	if strings.HasPrefix(relative, "http://") || strings.HasPrefix(relative, "https://") {
		return relative
	}

	// Handle absolute paths
	if strings.HasPrefix(relative, "/") {
		parts := strings.SplitN(base, "/", 4)
		if len(parts) >= 3 {
			return parts[0] + "//" + parts[2] + relative
		}
	}

	// Handle relative paths
	basePath := base[:strings.LastIndex(base, "/")+1]
	return basePath + relative
}

func (e *CrawlerEngine) GetJobStatus(jobID string) *CrawlJob {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.jobs[jobID]
}

func (e *CrawlerEngine) StopJob(jobID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if job, exists := e.jobs[jobID]; exists {
		job.Status = "stopped"
		job.EndTime = time.Now()
		return nil
	}
	return fmt.Errorf("job %s not found", jobID)
}

func (e *CrawlerEngine) GetResults(jobID string) []CrawlResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if job, exists := e.jobs[jobID]; exists {
		return job.Results
	}
	return nil
}
