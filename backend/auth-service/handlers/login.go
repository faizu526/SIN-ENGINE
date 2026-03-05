package handlers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/sin-engine/auth-service/models"
)

type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    AccessToken  string `json:"accessToken"`
    RefreshToken string `json:"refreshToken"`
    ExpiresIn    int64  `json:"expiresIn"`
    TokenType    string `json:"tokenType"`
    User         gin.H  `json:"user"`
}

func Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Find user
    var user models.User
    if err := db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }
    
    // Check password
    if !user.CheckPassword(req.Password) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }
    
    // Check if user is active
    if !user.Active {
        c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled"})
        return
    }
    
    // Generate tokens
    accessToken, refreshToken, err := GenerateTokens(&user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
        return
    }
    
    // Save session
    session := &models.Session{
        UserID:       user.ID,
        Token:        accessToken,
        RefreshToken: refreshToken,
        UserAgent:    c.Request.UserAgent(),
        ClientIP:     c.ClientIP(),
        ExpiresAt:    time.Now().Add(15 * time.Minute),
    }
    db.Create(session)
    
    // Update last login
    now := time.Now()
    user.LastLogin = &now
    db.Save(&user)
    
    c.JSON(http.StatusOK, LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    15 * 60,
        TokenType:    "Bearer",
        User: gin.H{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
            "fullName": user.FullName,
            "role":     user.Role,
        },
    })
}