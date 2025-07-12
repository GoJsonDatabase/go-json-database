package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ListRecord(c *gin.Context) {
	collection := c.Param("collection")
	dirPath := filepath.Join("database", collection)

	// check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read collection"})
		return
	}

	var items []map[string]interface{}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(dirPath, file.Name())

		f, err := os.Open(path)
		if err != nil {
			continue // skip unreadable files
		}

		var item map[string]interface{}
		if err := json.NewDecoder(f).Decode(&item); err == nil {
			items = append(items, item)
		}
		f.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"collection": collection,
		"count":      len(items),
		"items":      items,
	})
}
