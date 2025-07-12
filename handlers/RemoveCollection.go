package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func RemoveCollection(c *gin.Context) {
	var jsonData map[string]interface{}

	collectionId := c.Param("collection")

	// Parse JSON body
	if err := c.BindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	folderPath := "database/" + collectionId

	// Check if folder exists and is a directory
	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection folder does not exist"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Specified path is not a directory"})
		return
	}

	// Remove the folder and all its contents
	if err := os.RemoveAll(folderPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete collection folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Collection folder deleted successfully",
		"collection": collectionId,
	})
}
