package collectors

import "context"

// Collector defines the interface for breach data collectors
type Collector interface {
	GetName() string
	Search(ctx context.Context, query string) ([]Breach, error)
	IsAvailable() bool
}

// SearchResult combines results from multiple collectors
type SearchResult struct {
	Query     string   `json:"query"`
	Breaches  []Breach `json:"breaches"`
	Sources   []string `json:"sources"`
	Timestamp int64    `json:"timestamp"`
}

// SourceStatus represents the status of a data source
type SourceStatus struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	LastCheck int64  `json:"last_check"`
	Breaches  int    `json:"breaches_count"`
}
