package routes

import (
	"github.com/fahimaloy/blogo/controllers"
	"github.com/gin-gonic/gin"
)

func SetupBlogRoutes(router *gin.RouterGroup, bc *controllers.BlogController) {
	router.POST("/", bc.CreateBlog)
	router.GET("/", bc.GetAllPosts)
	router.GET("/:id", bc.GetBlogByID)
	router.PUT("/:id", bc.UpdateBlog)
	router.DELETE("/:id", bc.DeleteBlog)
}
