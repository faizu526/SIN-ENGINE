package handlers

import (
    "bytes"
    "io"
    "net/http"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

var serviceURLs = map[string]string{
    "auth-service":     "http://auth-service:8081",
    "scanner-service":  "http://scanner-service:8082",
    "crawler-service":  "http://crawler-service:8083",
    "search-service":   "http://search-service:8084",
    "exploit-service":  "http://exploitation-service:8085",
    "report-service":   "http://report-service:8086",
    "ai-assistant":     "http://ai-assistant:8087",
    "dork-service":     "http://dork-service:8088",
    "breach-service":   "http://data-breach:8089",
    "recon-service":    "http://recon-service:8090",
    "admin-service":    "http://admin-service:8091",
}

func ProxyTo(service, path string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get service URL
        baseURL, exists := serviceURLs[service]
        if !exists {
            c.JSON(http.StatusBadGateway, gin.H{"error": "service not found"})
            return
        }
        
        // Build target URL
        targetURL := baseURL + path
        
        // Replace path parameters
        for _, param := range c.Params {
            targetURL = strings.Replace(targetURL, ":"+param.Key, param.Value, -1)
        }
        
        // Add query parameters
        if c.Request.URL.RawQuery != "" {
            targetURL += "?" + c.Request.URL.RawQuery
        }
        
        // Create proxy request
        var body io.Reader
        if c.Request.Body != nil {
            bodyBytes, _ := io.ReadAll(c.Request.Body)
            body = bytes.NewReader(bodyBytes)
            c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
        }
        
        proxyReq, err := http.NewRequest(c.Request.Method, targetURL, body)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        // Copy headers
        proxyReq.Header = c.Request.Header.Clone()
        proxyReq.Header.Set("X-Request-ID", uuid.New().String())
        proxyReq.Header.Set("X-User-ID", c.GetString("user_id"))
        proxyReq.Header.Set("X-User-Role", c.GetString("role"))
        
        // Execute request
        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(proxyReq)
        if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
            return
        }
        defer resp.Body.Close()
        
        // Copy response headers
        for k, v := range resp.Header {
            c.Writer.Header()[k] = v
        }
        
        // Set status code
        c.Writer.WriteHeader(resp.StatusCode)
        
        // Copy response body
        io.Copy(c.Writer, resp.Body)
    }
}