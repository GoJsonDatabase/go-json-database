package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func UpdateRecord(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")

	// Read JSON data from the request body
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Build file path
	path := filepath.Join("database", collection, id+".json")

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println(path)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Marshal the new data to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not marshal JSON"})
		return
	}

	// Overwrite the file
	if err := ioutil.WriteFile(path, jsonBytes, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not write file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "updated",
		"id":         id,
		"collection": collection,
		"data":       data,
	})
}
