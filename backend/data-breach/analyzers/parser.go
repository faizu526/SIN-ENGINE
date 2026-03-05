package analyzers

import (
	"regexp"
	"strings"
	"time"
)

// Parser parses breach data
type Parser struct {
	emailRegex *regexp.Regexp
	dateRegex  *regexp.Regexp
}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{
		emailRegex: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		dateRegex:  regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
	}
}

// ParseEmail extracts email from text
func (p *Parser) ParseEmail(text string) []string {
	return p.emailRegex.FindAllString(text, -1)
}

// ParseDate extracts date from text
func (p *Parser) ParseDate(text string) []string {
	return p.dateRegex.FindAllString(text, -1)
}

// ParseBreachData parses raw breach data
func (p *Parser) ParseBreachData(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		// Normalize keys to lowercase
		normalizedKey := strings.ToLower(key)

		switch v := value.(type) {
		case string:
			result[normalizedKey] = strings.TrimSpace(v)
		case []interface{}:
			result[normalizedKey] = p.parseArray(v)
		default:
			result[normalizedKey] = value
		}
	}

	return result
}

func (p *Parser) parseArray(arr []interface{}) []string {
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// ExtractIOCs extracts Indicators of Compromise from breach data
func (p *Parser) ExtractIOCs(text string) map[string][]string {
	iocs := make(map[string][]string)

	// Extract emails
	emails := p.ParseEmail(text)
	if len(emails) > 0 {
		iocs["emails"] = emails
	}

	// Extract dates (potential breach dates)
	dates := p.ParseDate(text)
	if len(dates) > 0 {
		iocs["dates"] = dates
	}

	return iocs
}

// ParseTimestamp parses various timestamp formats
func (p *Parser) ParseTimestamp(value interface{}) time.Time {
	switch v := value.(type) {
	case string:
		// Try parsing common formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02",
			"02/01/2006",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
	case float64:
		return time.Unix(int64(v), 0)
	case int64:
		return time.Unix(v, 0)
	}
	return time.Now()
}
