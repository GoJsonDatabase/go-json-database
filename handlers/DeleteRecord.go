package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
)

func DeleteRecord(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")

	// Build the file path
	path := filepath.Join("database", collection, id+".json")

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Delete the file
	if err := os.Remove(path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "deleted",
		"id":         id,
		"collection": collection,
	})
}
