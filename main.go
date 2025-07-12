package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Extracted users map for authentication and response
var users = map[string]string{
	"foo": "bar",
}

// Extracted users map for authentication and response
var tokens = map[string]string{
	"foo": "JDJhJDEwJGt1cUs5eVprMUszdmlZVTlGWXZxSWV3dlUuM0RUcTM1dHlMWFRCNWtIQTZXeG9nRC5IUVdh",
}

func getItem(c *gin.Context) {
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

func createCollection(c *gin.Context) {
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

	basePath := "database"
	collectionPath := filepath.Join(basePath, name)

	if err := os.MkdirAll(collectionPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create collection: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Collection '%s' created at '%s'", name, collectionPath),
		"path":    collectionPath,
	})
}

func UpdateJSONFile(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")

	// Read JSON data from the request body
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Build file path
	path := filepath.Join("database", collection, id+".json")

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println(path)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Marshal the new data to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not marshal JSON"})
		return
	}

	// Overwrite the file
	if err := ioutil.WriteFile(path, jsonBytes, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not write file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "updated",
		"id":         id,
		"collection": collection,
		"data":       data,
	})
}

func CreateJSONFile(c *gin.Context) {
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

func adminHandler(c *gin.Context) {
	var req struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	pass, ok := users[req.Identity]
	if !ok || pass != req.Password {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	secretPassword := []byte(uuid.NewString())

	// Hash the password
	hashed, err := bcrypt.GenerateFromPassword(secretPassword, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hashed password: %s\n", hashed)

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"token":  hashed,
		"user":   req.Identity,
	})
}

func adminCheckHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Token    string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	token, ok := tokens[req.Username]
	fmt.Println(token)
	if !ok || token != req.Token {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"user":   req.Token,
	})
}

func getCollection(c *gin.Context) {
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

func renameCollectionFolder(collectionID, name string) error {
	oldPath := "database/" + collectionID
	newPath := "database/" + name

	// Check if oldPath exists and is a directory
	info, err := os.Stat(oldPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("source folder '%s' does not exist", oldPath)
	}
	if err != nil {
		return fmt.Errorf("error checking source folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source path '%s' is not a directory", oldPath)
	}

	// Attempt to rename
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("error renaming folder: %w", err)
	}

	return nil
}

func UpdateCollectionName(c *gin.Context) {
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

	if err := renameCollectionFolder(collectionID, name); err != nil {
		fmt.Println("Error:", err)
		c.Status(404)
		return
	}
	fmt.Println("Folder renamed successfully!")

	c.JSON(http.StatusOK, gin.H{
		"message":       "Collection name updated successfully",
		"collection_id": collectionID,
		"updated_name":  name,
	})
}

func RemoveCollectionFolder(c *gin.Context) {
	var jsonData map[string]interface{}

	collectionId := c.Param("collection")

	// Parse JSON body
	if err := c.BindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	folderPath := "database/" + collectionId

	// Check if folder exists and is a directory
	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection folder does not exist"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Specified path is not a directory"})
		return
	}

	// Remove the folder and all its contents
	if err := os.RemoveAll(folderPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete collection folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Collection folder deleted successfully",
		"collection": collectionId,
	})
}

func listCollection(c *gin.Context) {
	databasePath := "database"

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

func itemDelete(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")

	// Build the file path
	path := filepath.Join("database", collection, id+".json")

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Delete the file
	if err := os.Remove(path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "deleted",
		"id":         id,
		"collection": collection,
	})
}

func main() {
	r := gin.Default()
	// CRUD Record
	// list Record with Filter ?
	r.GET("/collection/:collection/:id", getItem)
	r.POST("/collection/:collection", CreateJSONFile)
	r.PATCH("/collection/:collection/:id", UpdateJSONFile)
	r.DELETE("/collection/:collection/:id", itemDelete)

	// CRUD Collection
	r.GET("/collections", listCollection)
	r.GET("/collections/:collection", getCollection)
	r.POST("/collections/", createCollection)
	r.PATCH("/collections/:collection", UpdateCollectionName)
	r.DELETE("/collections/:collection", RemoveCollectionFolder)

	// SuperUser
	r.POST("/admin/login", adminHandler)
	r.POST("/admin/check", adminCheckHandler)

	// Auth Support
	// File Upload
	// S3 Support
	// Mail Support

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
