package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Monitor represents a breach monitoring entry
type Monitor struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateMonitor creates a new breach monitor
func CreateMonitor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database not available",
			})
			return
		}

		var request struct {
			Email  string `json:"email" binding:"required,email"`
			UserID uint   `json:"user_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		monitor := Monitor{
			Email:  request.Email,
			UserID: request.UserID,
		}

		if err := db.Create(&monitor).Error; err != nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Monitor already exists for this email",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Monitor created successfully",
			"monitor": monitor,
		})
	}
}

// ListMonitors lists all monitors for a user
func ListMonitors(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database not available",
			})
			return
		}

		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "user_id parameter is required",
			})
			return
		}

		var monitors []Monitor
		if err := db.Where("user_id = ?", userID).Find(&monitors).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch monitors",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"monitors": monitors,
			"total":    len(monitors),
		})
	}
}

// DeleteMonitor deletes a monitor
func DeleteMonitor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database not available",
			})
			return
		}

		monitorID := c.Param("id")
		var monitor Monitor

		if err := db.First(&monitor, monitorID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Monitor not found",
			})
			return
		}

		if err := db.Delete(&monitor).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete monitor",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Monitor deleted successfully",
		})
	}
}
