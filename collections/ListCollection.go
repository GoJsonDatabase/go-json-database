package collections

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func ListCollection(c *gin.Context) {
	databasePath := "database"

	// auth.CheckAuth(c)

	entries, err := os.ReadDir(databasePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var collections []string

	for _, entry := range entries {
		if entry.IsDir() {
			collections = append(collections, entry.Name())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"collections": collections,
	})
}
