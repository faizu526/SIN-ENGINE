package main

import (
    "log"
    "os"
    
    "github.com/gin-gonic/gin"
    "github.com/sin-engine/auth-service/handlers"
    "github.com/sin-engine/auth-service/middleware"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var db *gorm.DB

func main() {
    // Initialize database
    dsn := "host=postgres user=sin_admin password=SinEngine@123 dbname=sin_engine port=5432 sslmode=disable"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Auto migrate
    db.AutoMigrate(&models.User{}, &models.Session{})
    
    // Set up router
    r := gin.Default()
    r.Use(middleware.Logger())
    
    // Routes
    auth := r.Group("/auth")
    {
        auth.POST("/login", handlers.Login)
        auth.POST("/register", handlers.Register)
        auth.POST("/refresh", handlers.Refresh)
        auth.POST("/logout", middleware.Auth(), handlers.Logout)
        auth.GET("/me", middleware.Auth(), handlers.GetMe)
        auth.PUT("/profile", middleware.Auth(), handlers.UpdateProfile)
        auth.POST("/change-password", middleware.Auth(), handlers.ChangePassword)
    }
    
    // Admin routes
    admin := r.Group("/admin")
    admin.Use(middleware.Auth(), middleware.AdminOnly())
    {
        admin.GET("/users", handlers.GetUsers)
        admin.GET("/users/:id", handlers.GetUser)
        admin.PUT("/users/:id", handlers.UpdateUser)
        admin.DELETE("/users/:id", handlers.DeleteUser)
    }
    
    log.Println("Auth service starting on port 8081")
    r.Run(":8081")
}