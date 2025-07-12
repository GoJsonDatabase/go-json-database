package handlers

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

func GetCollection(c *gin.Context) {
	collection := c.Param("collection")
	dir := filepath.Join("database", collection)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	total := len(files)
	if total < 0 {
		total = 0 // prevent negative count if directory is empty
	}

	c.JSON(http.StatusOK, gin.H{
		"collection": collection,
		"count":      total,
	})
}
