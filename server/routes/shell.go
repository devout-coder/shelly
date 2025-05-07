package routes

import (
	"fmt"
	"shelly/controllers"
	"shelly/middleware"

	"github.com/gin-gonic/gin"
)

func SetupShellRoutes(router *gin.Engine) {
	shell := router.Group("/shell")
	shell.Use(middleware.AuthMiddleware())
	fmt.Println("over here")
	shell.POST("", controllers.CreateShell)
	shell.DELETE("", controllers.DeleteShell)
	shell.GET("/ws", controllers.HandleShellWebSocket)
}
