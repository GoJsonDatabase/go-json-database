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
	r.GET("/api/collection/:collection", handlers.ListRecord)
	r.GET("/api/collection/:collection/:id", handlers.GetRecord)
	r.POST("/api/collection/:collection", handlers.CreateRecord)
	r.PATCH("/api/collection/:collection/:id", handlers.UpdateRecord)
	r.DELETE("/api/collection/:collection/:id", handlers.DeleteRecord)

	// CRUD Collection
	r.GET("/api/collections", handlers.ListCollection)
	r.GET("/api/collections/:collection", handlers.GetCollection)
	r.POST("/api/collections/", handlers.CreateCollection)
	r.PATCH("/api/collections/:collection", handlers.UpdateCollection)
	r.DELETE("/api/collections/:collection", handlers.RemoveCollection)

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
