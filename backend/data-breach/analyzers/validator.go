package analyzers

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator validates breach data
type Validator struct {
	emailRegex  *regexp.Regexp
	domainRegex *regexp.Regexp
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		emailRegex:  regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		domainRegex: regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?\.[a-zA-Z]{2,}$`),
	}
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(email string) bool {
	return v.emailRegex.MatchString(email)
}

// ValidateDomain validates a domain
func (v *Validator) ValidateDomain(domain string) bool {
	return v.domainRegex.MatchString(domain)
}

// ValidateBreachData validates breach data structure
func (v *Validator) ValidateBreachData(data map[string]interface{}) error {
	// Check required fields
	requiredFields := []string{"name", "title", "domain"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate domain if present
	if domain, ok := data["domain"].(string); ok {
		if !v.ValidateDomain(domain) {
			return fmt.Errorf("invalid domain format: %s", domain)
		}
	}

	return nil
}

// NormalizeDataClass normalizes data class names
func (v *Validator) NormalizeDataClass(dataClass string) string {
	// Map common variations to standard names
	normalizationMap := map[string]string{
		"email":       "Email addresses",
		"email_addr":  "Email addresses",
		"emailaddr":   "Email addresses",
		"pass":        "Passwords",
		"passwords":   "Passwords",
		"pwd":         "Passwords",
		"phone":       "Phone numbers",
		"phonenumber": "Phone numbers",
		"mobile":      "Phone numbers",
		"name":        "Names",
		"fullname":    "Names",
		"username":    "Usernames",
		"user":        "Usernames",
		"ip":          "IP addresses",
		"ipaddress":   "IP addresses",
		"address":     "Physical addresses",
		"city":        "Cities",
	}

	normalized := strings.ToLower(strings.TrimSpace(dataClass))
	if standard, ok := normalizationMap[normalized]; ok {
		return standard
	}

	// Capitalize first letter
	return strings.Title(dataClass)
}

// ValidateSearchQuery validates a search query
func (v *Validator) ValidateSearchQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	if len(query) < 3 {
		return fmt.Errorf("search query must be at least 3 characters")
	}

	if len(query) > 256 {
		return fmt.Errorf("search query must be less than 256 characters")
	}

	return nil
}

// CalculateRiskScore calculates a risk score based on breach data
func (v *Validator) CalculateRiskScore(dataClasses []string) int {
	// Higher risk data types
	highRisk := map[string]int{
		"passwords":               10,
		"credit cards":            10,
		"bank account numbers":    10,
		"social security numbers": 10,
		"national id numbers":     10,
	}

	mediumRisk := map[string]int{
		"email addresses":    5,
		"phone numbers":      5,
		"physical addresses": 5,
		"dates of birth":     5,
		"usernames":          4,
	}

	lowRisk := map[string]int{
		"ip addresses":       2,
		"browser details":    1,
		"device information": 1,
	}

	score := 0
	for _, dc := range dataClasses {
		normalized := strings.ToLower(dc)

		if points, ok := highRisk[normalized]; ok {
			score += points
		} else if points, ok := mediumRisk[normalized]; ok {
			score += points
		} else if points, ok := lowRisk[normalized]; ok {
			score += points
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}
