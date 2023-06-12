package controllers

import (
	"context"
	"errors"

	// "encoding/json"
	// "log"
	"net/http"
	"time"

	// "errors"

	"github.com/fahimaloy/blogo/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	Collection *mongo.Collection
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func (uc *UserController) Register(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = string(hashedPassword)

	_, err = uc.Collection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}
func (uc *UserController) Me(c *gin.Context) {
	username, err := getAuthUsername(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	c.JSON(http.StatusOK, username)
}

func getAuthUsername(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header is empty")
	}

	// Extract the token from the "Authorization" header
	tokenStr := authHeader[len("Bearer "):]

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Replace "your-secret-key" with your own secret key
		return []byte("your-secret-key"), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", errors.New("Invalid token signature")
		}
		return "", errors.New("Invalid token")
	}

	// Verify the token and retrieve the user ID
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		userID := claims.Issuer
		if userID == "" {
			return "", errors.New("Invalid claim: user_id")
		}
		return userID, nil
	}

	return "", errors.New("Invalid token")
}

func (uc *UserController) Login(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storedUser := models.User{}
	err = uc.Collection.FindOne(context.Background(), bson.M{"username": user.Username}).Decode(&storedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func generateToken(username string) (string, error) {
	claims := Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
			IssuedAt:  time.Now().Unix(),
			Subject:   "authentication",
			Issuer:    username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key")) // Replace with your own secret key
}

func (uc *UserController) SeedUser(c *gin.Context) {
	username := c.Param("username")
	password := c.Param("password")

	if username == "fahimaloy" && password == "32446+fa" {

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		user := models.User{
			Username: username,
			Password: string(hashedPassword),
			Role:     "superuser",
		}

		_, err = uc.Collection.InsertOne(context.Background(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User seeded successfully"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}
