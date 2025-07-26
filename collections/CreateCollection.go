package collections

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
	id := uuid.New()

	basePath := "database"
	collectionPath := filepath.Join(basePath, id.String())
	filePath := collectionPath + "/config.json"

	if err := os.MkdirAll(collectionPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create collection: %v", err)})
		return
	}

	// Open the file for writing
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	// Encode content as pretty JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	now := time.Now()
	germanDate := now.Format("02.01.2006 15:04")

	content := map[string]interface{}{
		"name":    name,
		"created": germanDate,
	}

	if err := encoder.Encode(content); err != nil {
		log.Fatalf("failed to write JSON: %v", err)
	}

	fmt.Printf("JSON file created: %s\n", filePath)

	c.JSON(http.StatusCreated, gin.H{
		"message":           fmt.Sprintf("Collection '%s' created at '%s'", name, filePath),
		"path":              collectionPath,
		"collection_config": filePath,
		"slug":              s,
	})
}
