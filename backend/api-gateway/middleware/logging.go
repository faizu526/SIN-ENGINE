package middleware

import (
    "time"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Process request
        c.Next()
        
        // Log after request
        duration := time.Since(start)
        
        logger.Info("API Request",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.String("ip", c.ClientIP()),
            zap.Duration("duration", duration),
            zap.String("user_agent", c.Request.UserAgent()),
        )
    }
}