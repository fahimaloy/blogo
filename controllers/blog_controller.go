package controllers

import (
	"context"
	"log"
	"net/http"

	// "time"
	"errors"

	"github.com/fahimaloy/blogo/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BlogController struct {
	Collection *mongo.Collection
}

func (bc *BlogController) GetBlogByID(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blog ID"})
		return
	}

	var blog models.Blog
	err = bc.Collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&blog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Blog not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get blog"})
		return
	}

	c.JSON(http.StatusOK, blog)
}

func (bc *BlogController) GetAllPosts(c *gin.Context) {
	var blogs []models.Blog

	cursor, err := bc.Collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get posts"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var blog models.Blog
		err := cursor.Decode(&blog)
		if err != nil {
			log.Println("Failed to decode blog:", err)
			continue
		}
		blogs = append(blogs, blog)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get posts"})
		return
	}

	if blogs == nil {
		blogs = []models.Blog{} // Initialize as an empty array if it's nil
	}

	c.JSON(http.StatusOK, blogs)
}

func (bc *BlogController) CreateBlog(c *gin.Context) {
	var blog models.Blog
	err := c.ShouldBindJSON(&blog)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if title, content, and author are not empty
	if blog.Title == "" || blog.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, content, and author fields cannot be empty"})
		return
	}

	// Retrieve the authenticated user's ID from the token
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	blog.Author = userID

	_, err = bc.Collection.InsertOne(context.Background(), blog)
	if err != nil {
		log.Println("Failed to create blog:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blog"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Blog created successfully"})
}
func (bc *BlogController) UpdateBlog(c *gin.Context) {
	// Retrieve the authenticated user's ID from the token
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the blog ID from the request parameters
	blogID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blog ID"})
		return
	}

	var blog models.Blog
	err = c.ShouldBindJSON(&blog)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the blog by ID and author
	filter := bson.M{"_id": objectID, "author": userID}
	update := bson.M{"$set": bson.M{"title": blog.Title, "content": blog.Content}}

	// Perform the update operation
	result, err := bc.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("Failed to update blog:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blog"})
		return
	}

	// Check if any blog was updated
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog not found or you don't have permission to update it"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Blog updated successfully"})
}

func (bc *BlogController) DeleteBlog(c *gin.Context) {

	// Get the authenticated user ID

	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Get the blog ID from the request parameters
	blogID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(blogID)

	// Define the filter to find the blog by ID and author
	// Log the filter for debugging purposes
	filter := bson.M{"_id": objectID, "author": userID}
	// log.Println("Filter:", filter)
	// var blog models.Blog
	// log.Println(bc.Collection.FindOne(context.Background(), filter).Decode(&blog))
	// log.Println(bc.Collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&blog))
	// Delete the blog
	result, err := bc.Collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Println("Failed to delete blog:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blog"})
		return
	}

	// Check if any blog was deleted
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog not found or you don't have permission to delete it"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Blog deleted successfully"})
}

// func (bc *BlogController) DeleteBlog(c *gin.Context) {
// 	// Get the authenticated user ID
// 	userID, err := getAuthenticatedUserID(c)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 		return
// 	}

// 	// Get the blog ID from the request parameters
// 	blogID := c.Param("id")

// 	// Define the filter to find the blog by ID and author
// 	filter := bson.M{"_id": blogID, "author": userID}
// 	log.Println(filter)
// 	// Delete the blog
// 	result, err := bc.Collection.DeleteOne(context.Background(), filter)
// 	log.Println(result)
// 	if err != nil {
// 		log.Println("Failed to delete blog:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blog"})
// 		return
// 	}
// 	log.Println(result.DeletedCount)
// 	// Check if any blog was deleted
// 	if result.DeletedCount == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Blog not found or you don't have permission to delete it"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Blog deleted successfully"})
// }

func getAuthenticatedUserID(c *gin.Context) (string, error) {
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
