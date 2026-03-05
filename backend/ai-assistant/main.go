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
	"github.com/sin-engine/ai-assistant/agents"
	"github.com/sin-engine/ai-assistant/handlers"
	"github.com/sin-engine/ai-assistant/memory"
	"github.com/sin-engine/ai-assistant/tools"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting AI Assistant Service...")

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

	// Initialize memory systems
	shortTerm := memory.NewShortTermMemory(redisClient, 100) // Last 100 messages
	longTerm := memory.NewLongTermMemory(db, redisClient)    // Persistent storage

	// Initialize tools
	searchTool := tools.NewSearchTool(getEnv("SEARCH_SERVICE_URL", "http://localhost:8081"))
	dorkTool := tools.NewDorkTool(getEnv("DORK_SERVICE_URL", "http://localhost:8083"))
	scannerTool := tools.NewScannerTool(getEnv("SCANNER_SERVICE_URL", "http://localhost:8082"))
	crawlerTool := tools.NewCrawlerTool(getEnv("CRAWLER_SERVICE_URL", "http://localhost:8084"))
	breachTool := tools.NewBreachTool()
	exploiterTool := tools.NewExploiterTool()

	availableTools := []tools.Tool{
		searchTool,
		dorkTool,
		scannerTool,
		crawlerTool,
		breachTool,
		exploiterTool,
	}

	// Initialize agents
	analyzer := agents.NewAnalyzer()
	planner := agents.NewPlanner(availableTools)
	executor := agents.NewExecutor(availableTools, shortTerm, longTerm)

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
			"service":   "ai-assistant",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Chat endpoints
		api.POST("/chat", handlers.Chat(analyzer, planner, executor, shortTerm, longTerm))
		api.POST("/chat/stream", handlers.ChatStream(analyzer, planner, executor, shortTerm))

		// Command endpoints
		api.POST("/command", handlers.ExecuteCommand(executor))
		api.GET("/commands", handlers.ListCommands())

		// Task endpoints
		api.POST("/task", handlers.CreateTask(planner, executor))
		api.GET("/task/:id", handlers.GetTaskStatus(executor))
		api.GET("/tasks", handlers.ListTasks(db))

		// Memory endpoints
		api.GET("/memory/short", handlers.GetShortTermMemory(shortTerm))
		api.GET("/memory/long", handlers.GetLongTermMemory(longTerm))
		api.DELETE("/memory", handlers.ClearMemory(shortTerm, longTerm))

		// Tool endpoints
		api.GET("/tools", handlers.ListTools(availableTools))
		api.POST("/tools/:name/execute", handlers.ExecuteTool(availableTools))

		// Learning endpoints
		api.POST("/learn", handlers.Learn(db, longTerm))
		api.GET("/learn/:topic", handlers.GetKnowledge(longTerm))
	}

	// Start server
	port := getEnv("PORT", "8085")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("AI Assistant service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down AI Assistant service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	if redisClient != nil {
		redisClient.Close()
	}

	log.Println("AI Assistant service exited")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
