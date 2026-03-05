package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/data-breach/analyzers"
	"github.com/sin-engine/data-breach/collectors"
	"github.com/sin-engine/data-breach/handlers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Data Breach Service...")

	// Database configuration
	dsn := getEnv("DATABASE_URL", "host=localhost user=postgres password=secret dbname=sin_engine port=5432 sslmode=disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		db = nil
	}

	// Initialize collectors (breach data sources)
	hibpCollector := collectors.NewHIBPCollector()
	dehashedCollector := collectors.NewDehashedCollector()
	leakcheckCollector := collectors.NewLeakCheckCollector()
	snusbaseCollector := collectors.NewSnusbaseCollector()

	collectorsList := []collectors.Collector{
		hibpCollector,
		dehashedCollector,
		leakcheckCollector,
		snusbaseCollector,
	}

	// Initialize analyzers
	parser := analyzers.NewParser()
	validator := analyzers.NewValidator()

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "data-breach",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Search endpoints
		api.POST("/search", handlers.Search(collectorsList, parser, db))
		api.GET("/search/:email", handlers.GetSearchResult(db))
		
		// Monitor endpoints
		api.POST("/monitor", handlers.CreateMonitor(db))
		api.GET("/monitor", handlers.ListMonitors(db))
		api.DELETE("/monitor/:id", handlers.DeleteMonitor(db))
 Alert endpoints
			
		//	api.POST("/alert", handlers.CreateAlert(db))
		api.GET("/alerts", handlers.ListAlerts(db))
		
		// Source status
		api.GET("/sources", handlers.GetSourcesStatus(collectorsList))
	}

	// Start server
	port := getEnv("PORT", "8086")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Data Breach service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Data Breach service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Data Breach service exited")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
