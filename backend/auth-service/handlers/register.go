package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/sin-engine/auth-service/models"
)

type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    FullName string `json:"fullName" binding:"required"`
}

func Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Check if user exists
    var count int64
    db.Model(&models.User{}).Where("username = ? OR email = ?", req.Username, req.Email).Count(&count)
    if count > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
        return
    }
    
    // Create user
    user := &models.User{
        Username: req.Username,
        Email:    req.Email,
        FullName: req.FullName,
        Role:     "user",
        APIKey:   generateAPIKey(),
        Active:   true,
    }
    
    if err := user.SetPassword(req.Password); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }
    
    if err := db.Create(user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "id":       user.ID,
        "username": user.Username,
        "email":    user.Email,
        "fullName": user.FullName,
        "role":     user.Role,
    })
}

func generateAPIKey() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}