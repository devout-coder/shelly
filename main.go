package main

import (
	"log"

	"shelly/config"
	"shelly/controllers"
	"shelly/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	config.ConnectDB()

	controllers.InitCollections()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Shelly",
		})
	})

	routes.AuthRoutes(r)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	} else {
		log.Println("Server started on port 8080")
	}
}
