package querier

import (
	"strings"
)

// QueryParser parses and normalizes search queries
type QueryParser struct{}

// NewQueryParser creates a new query parser
func NewQueryParser() *QueryParser {
	return &QueryParser{}
}

// Parse parses a search query into components
func (p *QueryParser) Parse(query string) *ParsedQuery {
	// Normalize query
	normalized := strings.ToLower(strings.TrimSpace(query))

	// Extract search terms
	terms := strings.Fields(normalized)

	// Extract filters
	filters := p.extractFilters(terms)

	// Remove filters from terms
	terms = p.removeFilters(terms, filters)

	// Extract sort order
	sortBy := "relevance"
	if contains(terms, "latest", "new", "recent") {
		sortBy = "date"
	} else if contains(terms, "popular", "top", "best") {
		sortBy = "popularity"
	}

	return &ParsedQuery{
		Original: query,
		Terms:    terms,
		Filters:  filters,
		SortBy:   sortBy,
		Page:     1,
		Limit:    20,
	}
}

// ParsedQuery represents a parsed search query
type ParsedQuery struct {
	Original string
	Terms    []string
	Filters  map[string]string
	SortBy   string
	Page     int
	Limit    int
}

// extractFilters extracts filter key-value pairs from terms
func (p *QueryParser) extractFilters(terms []string) map[string]string {
	filters := make(map[string]string)

	filterKeywords := map[string]string{
		"difficulty": "difficulty",
		"level":      "difficulty",
		"category":   "category",
		"type":       "type",
		"tag":        "tag",
		"author":     "author",
	}

	for i, term := range terms {
		for keyword, filterName := range filterKeywords {
			if strings.HasPrefix(term, keyword+":") {
				value := strings.TrimPrefix(term, keyword+":")
				filters[filterName] = value
				terms[i] = "" // Mark as used
				break
			}
		}
	}

	return filters
}

// removeFilters removes filter terms from the terms slice
func (p *QueryParser) removeFilters(terms []string, filters map[string]string) []string {
	var clean []string
	for _, term := range terms {
		if term != "" {
			clean = append(clean, term)
		}
	}
	return clean
}

// contains checks if any of the needles are in the haystack
func contains(haystack []string, needles ...string) bool {
	for _, needle := range needles {
		for _, hay := range haystack {
			if hay == needle {
				return true
			}
		}
	}
	return false
}

// ExpandQuery expands a query with synonyms
func (p *QueryParser) ExpandQuery(query string) []string {
	synonyms := map[string][]string{
		"hack":      {"penetration", "exploit", "security"},
		"web":       {"website", "webapp", "application"},
		"network":   {"networking", "infrastructure"},
		"password":  {"credential", "passwd", "pwd"},
		"injection": {"sqli", "sql", "nosql"},
		"xss":       {"cross-site", "scripting"},
		"csrf":      {"cross-site", "request", "forgery"},
	}

	terms := strings.Fields(strings.ToLower(query))
	var expanded []string

	for _, term := range terms {
		expanded = append(expanded, term)
		if syns, ok := synonyms[term]; ok {
			expanded = append(expanded, syns...)
		}
	}

	return expanded
}
