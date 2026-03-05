package middleware

import (
    "net/http"
    "sync"
    
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

var (
    userLimiters = make(map[string]*rate.Limiter)
    mu           sync.RWMutex
)

func getUserLimiter(userID string) *rate.Limiter {
    mu.RLock()
    limiter, exists := userLimiters[userID]
    mu.RUnlock()
    
    if !exists {
        limiter = rate.NewLimiter(10, 20) // 10 requests/sec, burst 20
        mu.Lock()
        userLimiters[userID] = limiter
        mu.Unlock()
    }
    
    return limiter
}

func RateLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            c.Next()
            return
        }
        
        limiter := getUserLimiter(userID)
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error":       "Rate limit exceeded",
                "retry_after": "1s",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}