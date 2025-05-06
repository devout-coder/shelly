package routes

import (
	"shelly/controllers"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	auth.POST("/signup", controllers.Signup)
	auth.POST("/login", controllers.Login)
}
