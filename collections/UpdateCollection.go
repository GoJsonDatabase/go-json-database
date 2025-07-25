package collections

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func UpdateCollection(c *gin.Context) {
	id := c.Param("collection")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	dir := "database/" + id

	// check if dir exists
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		fmt.Printf("directory %s does not exist\n", dir)
		os.Exit(1) // or return error if inside a function
	}

	// check if it's really a directory
	if !info.IsDir() {
		fmt.Printf("%s exists but is not a directory\n", dir)
		os.Exit(1)
	}

	// parse JSON body into a map
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid JSON: %v", err)})
		return
	}

	// ensure database/{id} directory exists
	dir = filepath.Join("database", id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("could not create directory: %v", err)})
		return
	}

	// file path
	filePath := filepath.Join(dir, "config.json")

	// marshal the original body
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to encode JSON: %v", err)})
		return
	}

	// append German date just before writing
	now := time.Now()
	germanDate := now.Format("02.01.2006 15:04:05")
	body["created"] = germanDate

	// re-marshal body with germanDate
	data, err = json.MarshalIndent(body, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to encode JSON with date: %v", err)})
		return
	}

	// write JSON to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to write file: %v", err)})
		return
	}

	// success
	c.JSON(http.StatusOK, gin.H{
		"message": "collection updated",
		"path":    filePath,
		"date":    germanDate,
	})
}
