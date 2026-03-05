package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sin-engine/data-breach/collectors"
	"gorm.io/gorm"
)

// Alert represents a breach alert
type Alert struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	UserID     uint   `gorm:"not null" json:"user_id"`
	Email      string `gorm:"not null" json:"email"`
	BreachName string `gorm:"not null" json:"breach_name"`
	Source     string `gorm:"not null" json:"source"`
	Seen       bool   `gorm:"default:false" json:"seen"`
}

// ListAlerts lists all alerts for a user
func ListAlerts(db *gorm.DB) gin.HandlerFunc {
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

		var alerts []Alert
		if err := db.Where("user_id = ?", userID).Order("id DESC").Find(&alerts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch alerts",
			})
			return
		}

		// Count unseen alerts
		var unseenCount int64
		db.Model(&Alert{}).Where("user_id = ? AND seen = ?", userID, false).Count(&unseenCount)

		c.JSON(http.StatusOK, gin.H{
			"alerts": alerts,
			"total":  len(alerts),
			"unseen": unseenCount,
		})
	}
}

// GetSourcesStatus returns the status of all breach data sources
func GetSourcesStatus(collectorsList []collectors.Collector) gin.HandlerFunc {
	return func(c *gin.Context) {
		statuses := make([]collectors.SourceStatus, 0)

		for _, collector := range collectorsList {
			status := collectors.SourceStatus{
				Name:      collector.GetName(),
				Available: collector.IsAvailable(),
				LastCheck: 0,
				Breaches:  0,
			}
			statuses = append(statuses, status)
		}

		c.JSON(http.StatusOK, gin.H{
			"sources": statuses,
			"total":   len(statuses),
		})
	}
}
