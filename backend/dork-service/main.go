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
	"github.com/sin-engine/dork-service/engines"
	"github.com/sin-engine/dork-service/handlers"
	"github.com/sin-engine/dork-service/scrapers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Dork Service...")

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

	// Initialize search engines
	googleEngine := engines.NewGoogleEngine()
	shodanEngine := engines.NewShodanEngine(getEnv("SHODAN_API_KEY", ""))
	censysEngine := engines.NewCensysEngine()
	fofaEngine := engines.NewFOFAEngine()
	githubEngine := engines.NewGitHubEngine()

	// Initialize scrapers
	exploitDBScraper := scrapers.NewExploitDBScraper()
	pastebinScraper := scrapers.NewPastebinScraper()

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
			"service":   "dork-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Dork search endpoints
		api.POST("/dork/search", handlers.SearchDork(googleEngine, shodanEngine, censysEngine, fofaEngine, redisClient))
		api.GET("/dork/google", handlers.SearchGoogle(googleEngine))
		api.GET("/dork/shodan", handlers.SearchShodan(shodanEngine))
		api.GET("/dork/censys", handlers.SearchCensys(censysEngine))
		api.GET("/dork/fofa", handlers.SearchFOFA(fofaEngine))
		api.GET("/dork/github", handlers.SearchGitHub(githubEngine))

		// Scraper endpoints
		api.GET("/exploitdb/search", handlers.SearchExploitDB(exploitDBScraper))
		api.GET("/pastebin/search", handlers.SearchPastebin(pastebinScraper))

		// Category endpoints
		api.GET("/dork/categories", handlers.GetCategories())
		api.GET("/dork/popular", handlers.GetPopularDorks(redisClient))
		api.POST("/dork/save", handlers.SaveDork(db))
		api.GET("/dork/history", handlers.GetDorkHistory(db))
	}

	// Start server
	port := getEnv("PORT", "8083")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Dork service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down dork service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	if redisClient != nil {
		redisClient.Close()
	}

	log.Println("Dork service exited")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
