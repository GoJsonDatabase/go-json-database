package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

// Extracted users map for authentication and response
var users = map[string]string{
	"foo": "bar",
}

func checkAuth(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return "", false
	}

	username, ok := users[authHeader]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return "", false
	}

	// Optional: Benutzer im Context speichern
	c.Set("username", username)
	return username, true
}

func ListCollection(c *gin.Context) {
	databasePath := "database"

	_, ok := checkAuth(c)
	if !ok {
		return
	}

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
