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
	"github.com/sin-engine/api-gateway/middleware"
	"github.com/sin-engine/api-gateway/routes"
	"github.com/sin-engine/pkg/database"
	migrations "github.com/sin-engine/pkg/database/migrations"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize database
	log.Println("🔄 Connecting to database...")
	if err := database.InitDB(); err != nil {
		log.Printf("⚠️  Database connection failed: %v", err)
		log.Println("⚠️  Running without database - some features may not work")
	} else {
		log.Println("✅ Database connected successfully")

		// Run migrations
		if err := migrations.RunMigrations(database.DB); err != nil {
			log.Printf("⚠️  Migration failed: %v", err)
		} else {
			log.Println("✅ Migrations completed")
		}
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimiter())

	// Setup routes
	routes.SetupRoutes(router)

	// Create server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("API Gateway starting on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
