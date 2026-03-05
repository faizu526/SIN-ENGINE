package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPastebinSources returns available pastebin sources
func GetPastebinSources() gin.HandlerFunc {
	return func(c *gin.Context) {
		sources := []map[string]string{
			{"id": "pastebin", "name": "Pastebin", "url": "https://pastebin.com"},
			{"id": "gist", "name": "GitHub Gist", "url": "https://gist.github.com"},
			{"id": "pasteee", "name": "Paste.ee", "url": "https://paste.ee"},
			{"id": "hastebin", "name": "Hastebin", "url": "https://hastebin.com"},
		}
		c.JSON(http.StatusOK, gin.H{"sources": sources})
	}
}
