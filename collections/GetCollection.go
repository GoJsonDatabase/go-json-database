package collections

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetCollection(c *gin.Context) {
	collection := c.Param("collection")
	dirPath := filepath.Join("database", collection)
	configPath := filepath.Join(dirPath, "config.json")

	// read config.json
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection config not found"})
		return
	}

	// parse JSON
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid config.json"})
		return
	}

	// get "config" key
	config, ok := raw["config"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": `"config" key not found in config.json`})
		return
	}

	// count JSON files (excluding config.json)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read collection directory"})
		return
	}

	count := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasSuffix(name, ".json") && name != "config.json" {
			count++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"collection": collection,
		"config":     config,
		"count":      count,
	})
}
