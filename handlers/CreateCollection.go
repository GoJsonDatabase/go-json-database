package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"net/http"
	"os"
	"path/filepath"
)

func CreateCollection(c *gin.Context) {
	var body map[string]interface{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	name, ok := body["name"].(string)
	if !ok || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'name' field"})
		return
	}
	s := slug.Make(name)

	basePath := "database"
	collectionPath := filepath.Join(basePath, s)

	if err := os.MkdirAll(collectionPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create collection: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Collection '%s' created at '%s'", name, collectionPath),
		"path":    collectionPath,
		"slug":    s,
	})
}
