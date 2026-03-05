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
	"github.com/redis/go-redis/v9"
	"github.com/sin-engine/crawler-service/engine"
	"github.com/sin-engine/crawler-service/handlers"
	"github.com/sin-engine/crawler-service/queue"
	"github.com/sin-engine/crawler-service/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Crawler Service...")

	// Database configuration
	dsn := getEnv("DATABASE_URL", "host=localhost user=postgres password=secret dbname=sin_engine port=5432 sslmode=disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		db = nil
	}

	// Redis configuration
	redisAddr := getEnv("REDIS_URL", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		redisClient = nil
	}

	// Initialize components
	crawlerEngine := engine.NewCrawlerEngine(db, redisClient)
	jobQueue := queue.NewJobQueue(redisClient)
	store := storage.NewStorage(db, redisClient)
	scheduler := engine.NewScheduler(jobQueue, crawlerEngine)

	// Start scheduler
	go scheduler.Start(ctx)

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
			"service":   "crawler-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Crawl endpoints
		api.POST("/crawl/start", handlers.StartCrawl(crawlerEngine, jobQueue))
		api.POST("/crawl/stop", handlers.StopCrawl(crawlerEngine))
		api.GET("/crawl/status/:job_id", handlers.GetCrawlStatus(crawlerEngine))
		api.GET("/crawl/results/:job_id", handlers.GetCrawlResults(store))

		// Queue endpoints
		api.GET("/queue/stats", handlers.GetQueueStats(jobQueue))
		api.POST("/queue/pause", handlers.PauseQueue(jobQueue))
		api.POST("/queue/resume", handlers.ResumeQueue(jobQueue))
	}

	// Start server
	port := getEnv("PORT", "8082")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Crawler service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down crawler service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Cleanup
	scheduler.Stop()
	if redisClient != nil {
		redisClient.Close()
	}

	log.Println("Crawler service exited")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
