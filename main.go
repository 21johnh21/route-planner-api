package main

import (
	"log"
	"os"

	"route-planner-api/config"
	"route-planner-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load DB and configs
	db := config.InitDB()
	defer db.Close()

	r := gin.Default()

	// Attach DB to context
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Setup routes
	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	r.Run(":" + port)
}
