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

	config.InitCollections()

	if err := controllers.InitKubernetesClient(); err != nil {
		log.Fatal("Failed to initialize Kubernetes client:", err)
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Shelly",
		})
	})

	routes.SetupAuthRoutes(r)
	routes.SetupShellRoutes(r)

	if err := r.Run(":8000"); err != nil {
		log.Fatal("Failed to start server:", err)
	} else {
		log.Println("Server started on port 8000")
	}
}
