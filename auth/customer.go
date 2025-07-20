package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func CustomerHandler(c *gin.Context) {
	users, err := loadSuperUsers("auth/superusers.json")
	if err != nil {
		log.Printf("failed to load superusers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var req struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON 1"})
		return
	}

	user, ok := users[req.Identity]
	if !ok || user.Password != req.Password {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// generate a new random token (plain)
	secretToken := []byte(uuid.NewString())

	// hash the token for storage
	hashed, err := bcrypt.GenerateFromPassword(secretToken, bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// load existing tokens
	tokens, err := loadTokens("auth/token.json")
	if err != nil {
		log.Printf("failed to load tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// append the new hashed token for this user
	userTokens := tokens[user.Identity]

	// limit to max 5 tokens, drop oldest if needed
	if len(userTokens) >= 5 {
		userTokens = userTokens[1:] // drop the oldest token
	}
	userTokens = append(userTokens, string(hashed))
	tokens[user.Identity] = userTokens

	// save updated tokens back to file
	if err := saveTokens("auth/token.json", tokens); err != nil {
		log.Printf("failed to save tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	fmt.Printf("New token for user %s: %s\n", user.Identity, hashed)

	// return the plain token (client uses this to authenticate later)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"token":  string(secretToken),
		"user":   user.Identity,
		"name":   user.Name,
	})
}

func CustomerCheckHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Token    string `json:"token"` // plain token sent by client
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON 2"})
		return
	}

	tokens, err := loadTokens("auth/token.json")
	if err != nil {
		log.Printf("failed to load tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	userTokens, ok := tokens[req.Username]
	if !ok || len(userTokens) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// check if any stored hashed token matches the provided plain token
	authenticated := false
	for _, hashedToken := range userTokens {
		err := bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(req.Token))
		if err == nil {
			authenticated = true
			break
		}
	}

	if !authenticated {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"user":   req.Username,
	})
}
