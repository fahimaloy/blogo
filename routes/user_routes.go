package routes

import (
	"github.com/fahimaloy/blogo/controllers"
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(router *gin.RouterGroup, uc *controllers.UserController) {
	router.POST("/register", uc.Register)
	router.POST("/login", uc.Login)
	router.GET("/seed/:username/:password", uc.SeedUser)
	router.GET("/me", uc.Me)
}
