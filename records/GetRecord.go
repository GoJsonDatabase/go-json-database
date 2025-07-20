package records

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
)

func GetRecord(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")

	path := filepath.Join("database", collection, id+".json")
	file, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON file"})
		return
	}
	c.JSON(http.StatusOK, data)
}
