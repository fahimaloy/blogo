package main

import (
	"context"
	"log"

	// "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fahimaloy/blogo/controllers"
	"github.com/fahimaloy/blogo/routes"
)

func main() {
	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("blogo")
	blogCollection := db.Collection("blogs")
	userCollection := db.Collection("users")

	blogController := &controllers.BlogController{
		Collection: blogCollection,
	}
	userController := &controllers.UserController{
		Collection: userCollection,
	}

	router := gin.Default()

	api := router.Group("/api")
	{
		userRoutes := api.Group("/users")
		{
			routes.SetupUserRoutes(userRoutes, userController)
		}

		blogRoutes := api.Group("/blogs")
		{
			blogRoutes.Use(authMiddleware) // Apply authentication middleware
			routes.SetupBlogRoutes(blogRoutes, blogController)
		}
	}

	router.Run(":8080")
}

func authMiddleware(c *gin.Context) {
	// Implement authentication logic
	// ...
	c.Next()
}
