package records

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func CreateRecord(c *gin.Context) {
	collection := c.Param("collection")
	id := uuid.NewString()

	// Read JSON data from the request body
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Make sure the collection directory exists
	if err := os.MkdirAll("database/"+collection, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create directory"})
		return
	}

	// Marshal the data to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not marshal JSON"})
		return
	}

	// Write to file
	path := filepath.Join("database/"+collection, id+".json")
	if err := ioutil.WriteFile(path, jsonBytes, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not write file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "created", "id": id, "collection": collection})
}
