package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/sin-engine/api-gateway/handlers"
    "github.com/sin-engine/api-gateway/middleware"
)

func SetupRoutes(r *gin.Engine) {
    // Health check
    r.GET("/health", handlers.HealthCheck)
    
    // API v1 group
    v1 := r.Group("/api/v1")
    {
        // Auth routes (no auth required)
        auth := v1.Group("/auth")
        {
            auth.POST("/login", handlers.ProxyTo("auth-service", "/login"))
            auth.POST("/register", handlers.ProxyTo("auth-service", "/register"))
            auth.POST("/refresh", handlers.ProxyTo("auth-service", "/refresh"))
        }
        
        // Protected routes
        protected := v1.Group("/")
        protected.Use(middleware.Auth())
        {
            // Auth protected
            protected.POST("/auth/logout", handlers.ProxyTo("auth-service", "/logout"))
            protected.GET("/auth/me", handlers.ProxyTo("auth-service", "/me"))
            
            // Scanner routes
            scanner := protected.Group("/scanner")
            {
                scanner.POST("/scan", handlers.ProxyTo("scanner-service", "/scan"))
                scanner.GET("/scans", handlers.ProxyTo("scanner-service", "/scans"))
                scanner.GET("/scan/:id", handlers.ProxyTo("scanner-service", "/scan/:id"))
                scanner.GET("/scan/:id/results", handlers.ProxyTo("scanner-service", "/scan/:id/results"))
            }
            
            // AI Assistant routes
            ai := protected.Group("/ai")
            {
                ai.POST("/chat", handlers.ProxyTo("ai-assistant", "/chat"))
                ai.POST("/command", handlers.ProxyTo("ai-assistant", "/command"))
            }
            
            // Recon routes
            recon := protected.Group("/recon")
            {
                recon.POST("/domain", handlers.ProxyTo("recon-service", "/domain"))
                recon.GET("/subdomains/:domain", handlers.ProxyTo("recon-service", "/subdomains/:domain"))
            }
            
            // Admin routes
            admin := protected.Group("/admin")
            admin.Use(middleware.AdminOnly())
            {
                admin.GET("/stats", handlers.ProxyTo("admin-service", "/stats"))
                admin.GET("/users", handlers.ProxyTo("auth-service", "/admin/users"))
            }
        }
    }
}