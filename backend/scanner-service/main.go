package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sin-engine/scanner-service/handlers"
    "github.com/sin-engine/scanner-service/middleware"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.uber.org/zap"
)

var (
    mongoClient *mongo.Client
    logger      *zap.Logger
)

func main() {
    // Initialize logger
    logger, _ = zap.NewProduction()
    defer logger.Sync()
    
    // Connect to MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb:27017"))
    if err != nil {
        logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
    }
    
    // Set Gin mode
    gin.SetMode(gin.ReleaseMode)
    
    // Create router
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.Logger(logger))
    r.Use(middleware.Auth())
    
    // Setup routes
    setupRoutes(r)
    
    // Create server
    srv := &http.Server{
        Addr:         ":8082",
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Start server
    go func() {
        logger.Info("Scanner service starting on port 8082")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Failed to start server", zap.Error(err))
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    logger.Info("Shutting down scanner service...")
    
    // Graceful shutdown
    ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        logger.Fatal("Server forced to shutdown", zap.Error(err))
    }
    
    // Disconnect MongoDB
    if err := mongoClient.Disconnect(ctx); err != nil {
        logger.Error("Failed to disconnect MongoDB", zap.Error(err))
    }
    
    logger.Info("Scanner service exited")
}

func setupRoutes(r *gin.Engine) {
    // Health check
    r.GET("/health", handlers.HealthCheck)
    
    // Scan routes
    scan := r.Group("/scan")
    {
        scan.POST("/", handlers.CreateScan)
        scan.GET("/list", handlers.ListScans)
        scan.GET("/:id", handlers.GetScan)
        scan.POST("/:id/start", handlers.StartScan)
        scan.POST("/:id/stop", handlers.StopScan)
        scan.GET("/:id/results", handlers.GetResults)
        scan.DELETE("/:id", handlers.DeleteScan)
    }
    
    // Vulnerability routes
    vuln := r.Group("/vulnerabilities")
    {
        vuln.GET("/", handlers.ListVulnerabilities)
        vuln.GET("/:id", handlers.GetVulnerability)
        vuln.POST("/:id/verify", handlers.VerifyVulnerability)
        vuln.POST("/:id/exploit", handlers.ExploitVulnerability)
    }
}