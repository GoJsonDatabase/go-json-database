package main

import (
	"github.com/gin-gonic/gin"
	"go-database-json/auth"
	"go-database-json/collections"
	"go-database-json/records"
)

// Extracted users map for authentication and response
var users = map[string]string{
	"foo": "bar",
}

// Extracted users map for authentication and response
var tokens = map[string]string{
	"foo": "JDJhJDEwJGt1cUs5eVprMUszdmlZVTlGWXZxSWV3dlUuM0RUcTM1dHlMWFRCNWtIQTZXeG9nRC5IUVdh",
}

func main() {
	r := gin.Default()

	// CRUD Record
	// TODO Filters for records ?

	collectiongroup := r.Group("/api/collection")
	{
		collectiongroup.GET("/:collection", records.ListRecord)
		collectiongroup.GET("/:collection/:id", records.GetRecord)
		collectiongroup.POST("/:collection", records.CreateRecord)
		collectiongroup.PATCH("/:collection/:id", records.UpdateRecord)
		collectiongroup.DELETE("/:collection/:id", records.DeleteRecord)
	}

	collectionsgroup := r.Group("/api/collections")
	{
		// CRUD Collection
		collectionsgroup.GET("/", collections.ListCollection)
		collectionsgroup.GET("/:collection", collections.GetCollection)
		collectionsgroup.POST("/", collections.CreateCollection)
		collectionsgroup.PATCH("/:collection", collections.UpdateCollection)
		collectionsgroup.DELETE("/:collection", collections.RemoveCollection)
	}

	// SuperUser
	superusergroup := r.Group("/api/superuser")
	{
		superusergroup.POST("/login", auth.AdminHandler)
		superusergroup.POST("/register", auth.RegisterHandler)
		superusergroup.POST("/check", auth.AdminCheckHandler)
	}

	// Customers
	customergroup := r.Group("/api/customer")
	{
		customergroup.POST("/login", auth.CustomerHandler)
		customergroup.POST("/register", auth.RegisterHandler)
		customergroup.POST("/check", auth.CustomerCheckHandler)
	}

	// File Upload
	// S3 Support
	// Mail Support

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
