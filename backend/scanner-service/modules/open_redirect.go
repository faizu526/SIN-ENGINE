package scanner

import (
    "fmt"
    "net/url"
    "strings"
    
    "github.com/sin-engine/scanner-service/models"
)

func (e *ScanEngine) testOpenRedirect() {
    params := []string{
        "next", "redirect", "redirect_uri", "redirect_to",
        "return_to", "return", "u", "url", "to", "destination",
        "continue", "follow", "callback", "referer", "ref",
        "source", "link", "goto", "target", "away",
    }
    
    payloads := []string{
        "https://evil.com",
        "//evil.com",
        "https://evil.com@target.com",
        "/\\evil.com",
        "https://evil.com:443",
        "https://evil.com#@target.com",
        "https://target.com.evil.com",
        "https://evil.com/target.com",
        "javascript:window.location='https://evil.com'",
    }
    
    // First, crawl the target to find URLs with parameters
    urls := e.crawlTarget()
    
    for _, targetURL := range urls {
        parsed, err := url.Parse(targetURL)
        if err != nil {
            continue
        }
        
        query := parsed.Query()
        
        for _, param := range params {
            for _, payload := range payloads {
                // Skip if parameter already exists with different value
                if _, exists := query[param]; exists {
                    continue
                }
                
                testURL := fmt.Sprintf("%s?%s=%s", targetURL, param, url.QueryEscape(payload))
                
                resp, err := e.client.Get(testURL)
                if err != nil {
                    continue
                }
                defer resp.Body.Close()
                
                location := resp.Header.Get("Location")
                if location != "" && strings.Contains(location, "evil.com") {
                    vuln := models.NewVulnerability()
                    vuln.Type = "Open Redirect"
                    vuln.Severity = models.SeverityHigh
                    vuln.Name = "Open Redirect Vulnerability"
                    vuln.URL = targetURL
                    vuln.Parameter = param
                    vuln.Payload = payload
                    vuln.Evidence = location
                    vuln.Description = "The application redirects to an external domain controlled by the attacker without proper validation."
                    vuln.Remediation = "Validate and whitelist redirect URLs. Avoid using user input in redirects."
                    vuln.CWE = "CWE-601"
                    vuln.CVSS = 6.5
                    
                    select {
                    case e.vulnChan <- vuln:
                    default:
                    }
                    
                    break // Found vulnerability, no need to test more payloads for this parameter
                }
            }
        }
    }
}

func (e *ScanEngine) crawlTarget() []string {
    // Simple crawling logic - in production, use the crawler service
    // This is a placeholder
    return []string{e.target}
}