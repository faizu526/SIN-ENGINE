package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "SIN-ENGINE API Gateway",
	})
}

// ProxyTo is defined in proxy.go
