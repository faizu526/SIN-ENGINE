package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/sin-engine/auth-service/models"
)

func GetUsers(c *gin.Context) {
    var users []models.User
    db.Find(&users)
    
    result := make([]gin.H, len(users))
    for i, user := range users {
        result[i] = gin.H{
            "id":        user.ID,
            "username":  user.Username,
            "email":     user.Email,
            "fullName":  user.FullName,
            "role":      user.Role,
            "active":    user.Active,
            "lastLogin": user.LastLogin,
            "createdAt": user.CreatedAt,
        }
    }
    
    c.JSON(http.StatusOK, result)
}

func GetUser(c *gin.Context) {
    id := c.Param("id")
    
    var user models.User
    if err := db.Where("id = ?", id).First(&user).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "id":        user.ID,
        "username":  user.Username,
        "email":     user.Email,
        "fullName":  user.FullName,
        "role":      user.Role,
        "active":    user.Active,
        "rateLimit": user.RateLimit,
        "apiKey":    user.APIKey,
        "lastLogin": user.LastLogin,
        "createdAt": user.CreatedAt,
    })
}

func UpdateUser(c *gin.Context) {
    id := c.Param("id")
    
    var updates map[string]interface{}
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Remove sensitive fields
    delete(updates, "id")
    delete(updates, "password_hash")
    delete(updates, "api_key")
    
    if err := db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
    id := c.Param("id")
    
    // Don't allow deleting yourself
    currentUserID := c.GetString("user_id")
    if currentUserID == id {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete yourself"})
        return
    }
    
    // Delete user
    if err := db.Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }
    
    // Delete sessions
    db.Where("user_id = ?", id).Delete(&models.Session{})
    
    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}