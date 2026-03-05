package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sin-engine/search-service/handlers"
	"github.com/sin-engine/search-service/indexer"
	"github.com/sin-engine/search-service/querier"
	"github.com/sin-engine/search-service/ranker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db           *gorm.DB
	redisClient  *redis.Client
	index        *indexer.SearchIndexer
	queryEngine  *querier.QueryEngine
	rankerEngine *ranker.RankerEngine
)

func main() {
	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Search Service...")

	// Initialize PostgreSQL
	dsn := "host=postgres user=sin_admin password=SinEngine@123 dbname=sin_engine port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Initialize Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		redisClient = nil
	} else {
		log.Println("Connected to Redis")
	}

	// Initialize search indexer
	index = indexer.NewSearchIndexer(db, redisClient)

	// Initialize query engine
	queryEngine = querier.NewQueryEngine(db, redisClient)

	// Initialize ranker
	rankerEngine = ranker.NewRankerEngine(db, redisClient)

	// Start background indexing
	go startBackgroundIndexing()

	// Set up Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "search",
		})
	})

	// Search routes
	search := r.Group("/api/search")
	{
		search.GET("", handlers.Search(queryEngine, rankerEngine))
		search.GET("/suggest", handlers.Suggest(queryEngine))
		search.POST("/filter", handlers.Filter())
		search.GET("/trending", handlers.Trending())
		search.GET("/recent", handlers.Recent())
	}

	// Admin routes for indexing
	admin := r.Group("/api/admin/search")
	{
		admin.POST("/reindex", handlers.Reindex(index))
		admin.GET("/stats", handlers.GetStats(index))
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down search service...")
		if redisClient != nil {
			redisClient.Close()
		}
		os.Exit(0)
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	log.Printf("Search service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startBackgroundIndexing() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Running background indexing...")
		if err := index.FullReindex(); err != nil {
			log.Printf("Background indexing error: %v", err)
		}
	}
}
