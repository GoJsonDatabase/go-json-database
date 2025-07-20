package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"strings"
)

type User struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Token    string `json:"token"`
}

func CheckAuth(c *gin.Context) (string, bool) {
	users, err := loadUsers("auth/superusers.json")
	if err != nil {
		log.Printf("failed to load users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return "", false
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return "", false
	}

	username, password, err := parseAuthHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return "", false
	}

	user, ok := users[username]
	fmt.Println(user)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return "", false
	}

	// compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return "", false
	}

	c.Set("username", user.Identity)
	return user.Identity, true
}

func loadUsers(path string) (map[string]User, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	users := make(map[string]User)
	if err := json.Unmarshal(bytes, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func parseAuthHeader(header string) (string, string, error) {
	// Expect "username:password"
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("Invalid Authorization format. Expected username:password")
	}
	return parts[0], parts[1], nil
}
