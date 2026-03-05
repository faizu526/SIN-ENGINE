package scanner

import (
    "fmt"
    "io"
    "net/url"
    "strings"
    
    "github.com/sin-engine/scanner-service/models"
)

func (e *ScanEngine) testXSS() {
    params := []string{"q", "s", "search", "query", "keyword", "term", "id", "page"}
    
    payloads := []string{
        "<script>alert('XSS')</script>",
        "<img src=x onerror=alert(1)>",
        "javascript:alert('XSS')",
        "\"><script>alert(1)</script>",
        "'><script>alert(1)</script>",
        "<svg/onload=alert(1)>",
        "<body onload=alert(1)>",
        "{{constructor.constructor('alert(1)')()}}",
        "${alert(1)}",
        "<iframe src=javascript:alert(1)>",
    }
    
    urls := e.crawlTarget()
    
    for _, targetURL := range urls {
        parsed, err := url.Parse(targetURL)
        if err != nil {
            continue
        }
        
        query := parsed.Query()
        
        for _, param := range params {
            for _, payload := range payloads {
                testURL := fmt.Sprintf("%s?%s=%s", targetURL, param, url.QueryEscape(payload))
                
                resp, err := e.client.Get(testURL)
                if err != nil {
                    continue
                }
                defer resp.Body.Close()
                
                body, _ := io.ReadAll(resp.Body)
                
                // Check if payload is reflected
                if strings.Contains(string(body), payload) {
                    vuln := models.NewVulnerability()
                    vuln.Type = "Cross-Site Scripting (XSS)"
                    vuln.Severity = models.SeverityHigh
                    vuln.Name = "Reflected XSS"
                    vuln.URL = targetURL
                    vuln.Parameter = param
                    vuln.Payload = payload
                    vuln.Evidence = fmt.Sprintf("Payload reflected in response: %s", payload)
                    vuln.Description = "User input is reflected in the response without proper sanitization, allowing execution of malicious scripts."
                    vuln.Remediation = "Implement proper input validation and output encoding. Use Content Security Policy (CSP) headers."
                    vuln.CWE = "CWE-79"
                    vuln.CVSS = 6.5
                    
                    select {
                    case e.vulnChan <- vuln:
                    default:
                    }
                    
                    break
                }
            }
        }
    }
}