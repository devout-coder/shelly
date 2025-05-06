package routes

import (
	"shelly/controllers"
	"shelly/middleware"

	"github.com/gin-gonic/gin"
)

func SetupShellRoutes(router *gin.Engine) {
	shell := router.Group("/shell")
	shell.Use(middleware.AuthMiddleware())
	shell.POST("/", controllers.CreateShell)
	shell.DELETE("/", controllers.DeleteShell)
}
