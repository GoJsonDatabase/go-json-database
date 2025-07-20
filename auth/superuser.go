package auth

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// TokenValid checks if the token of the user in the request matches the expected token
func TokenValid(c *gin.Context, expectedToken string) bool {
	var req map[string]User

	// Parse JSON body into map
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid JSON: %v", err),
		})
		return false
	}

	// Check that there is exactly one user in the request
	if len(req) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: expected exactly one user",
		})
		return false
	}

	// Extract the username and user info
	var username string
	var user User
	for k, v := range req {
		username = k
		user = v
		break
	}

	// Check if the provided token matches the expected one
	if user.Token != expectedToken {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("Unauthorized: invalid token for user %s", username),
		})
		return false
	}

	// Success
	c.Set("username", username)
	return true
}

var (
	// map username or IP to a rate limiter
	limiterStore = make(map[string]*rate.Limiter)
	mu           sync.Mutex
	rateLimit    = rate.Every(1 * time.Minute) // 1 request per second
	burstLimit   = 3                           // allow burst of 3 requests
)

func getLimiter(key string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := limiterStore[key]
	if !exists {
		limiter = rate.NewLimiter(rateLimit, burstLimit)
		limiterStore[key] = limiter
	}
	return limiter
}

func LoginHandler(c *gin.Context) {
	// use IP or username as key
	key := c.ClientIP()

	limiter := getLimiter(key)
	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Please try again later."})
		return
	}

	// ... your login logic here ...
}

type SuperUser struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func loadSuperUsers(path string) (map[string]SuperUser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var users map[string]SuperUser
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

// Tokens map: username -> slice of hashed tokens
func loadTokens(path string) (map[string][]string, error) {
	tokens := make(map[string][]string)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return tokens, nil
		}
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func saveTokens(path string, tokens map[string][]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(tokens)
}

func AdminHandler(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON 4"})
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

func AdminCheckHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Token    string `json:"token"` // plain token sent by client
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON 5"})
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

var req struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// RegisterHandler registers a new superuser with bcrypt-hashed password
func RegisterHandler(c *gin.Context) {

	const expectedToken = "token123"

	if !TokenValid(c, expectedToken) {
		return // response already sent
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	if req.Identity == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Identity and Password required"})
		return
	}

	// load existing superusers
	users := make(map[string]User)

	path := "auth/superusers.json"

	// try to read file if it exists
	if file, err := os.Open(path); err == nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&users); err != nil {
			log.Printf("failed to decode users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read users"})
			return
		}
	}

	// check if user already exists
	if _, exists := users[req.Identity]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// add new superuser
	users[req.Identity] = User{
		Identity: req.Identity,
		Password: string(hashedPassword),
	}

	// write back to file
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		log.Printf("failed to marshal users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("failed to write file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "created",
		"user":   req.Identity,
	})
}
