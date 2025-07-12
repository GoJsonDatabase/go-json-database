package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"net/http"
	"os"
)

func UpdateCollection(c *gin.Context) {
	collectionID := c.Param("collection")
	var jsonData map[string]interface{}

	if err := c.BindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	name, ok := jsonData["name"].(string)
	if !ok || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name field is required"})
		return
	}

	s := slug.Make(name)

	if err := renameFolder(collectionID, s); err != nil {
		fmt.Println("Error:", err)
		c.Status(404)
		return
	}
	fmt.Println("Folder renamed successfully!")

	c.JSON(http.StatusOK, gin.H{
		"message":       "Collection name updated successfully",
		"collection_id": collectionID,
		"updated_name":  s,
	})
}

func renameFolder(oldPath, newPath string) error {
	// Check if oldPath exists and is a directory
	info, err := os.Stat(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source folder %q does not exist", oldPath)
		}
		return fmt.Errorf("failed to check source folder %q: %w", oldPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source path %q is not a directory", oldPath)
	}

	// Check if newPath already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("target folder %q already exists", newPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check target folder %q: %w", newPath, err)
	}

	// Perform rename
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename folder %q to %q: %w", oldPath, newPath, err)
	}

	return nil
}
