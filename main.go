package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-database-json/handlers"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// Extracted users map for authentication and response
var users = map[string]string{
	"foo": "bar",
}

// Extracted users map for authentication and response
var tokens = map[string]string{
	"foo": "JDJhJDEwJGt1cUs5eVprMUszdmlZVTlGWXZxSWV3dlUuM0RUcTM1dHlMWFRCNWtIQTZXeG9nRC5IUVdh",
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

func main() {
	r := gin.Default()

	// CRUD Record
	// TODO Filters for records ?

	collection := r.Group("/api/collection")
	{
		collection.GET("/:collection", handlers.ListRecord)
		collection.GET("/:collection/:id", handlers.GetRecord)
		collection.POST("/:collection", handlers.CreateRecord)
		collection.PATCH("/:collection/:id", handlers.UpdateRecord)
		collection.DELETE("/:collection/:id", handlers.DeleteRecord)
	}

	collections := r.Group("/api/collections")
	{
		// CRUD Collection
		collections.GET("/", handlers.ListCollection)
		collections.GET("/:collection", handlers.GetCollection)
		collections.POST("/", handlers.CreateCollection)
		collections.PATCH("/:collection", handlers.UpdateCollection)
		collections.DELETE("/:collection", handlers.RemoveCollection)
	}

	// SuperUser
	r.POST("/api/admin/login", adminHandler)
	r.POST("/api/admin/check", adminCheckHandler)

	// Auth Support
	// File Upload
	// S3 Support
	// Mail Support

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
