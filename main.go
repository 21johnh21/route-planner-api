package main

import (
	"log"
	"os"

	"github.com/21johnh21/route-planner-api/config"
	"github.com/21johnh21/route-planner-api/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load DB and configs
	db := config.InitDB()
	sqlDB, err := db.DB() // get underlying *sql.DB
	if err != nil {
		log.Fatal("Failed to get sql.DB from GORM:", err)
	}
	defer sqlDB.Close()

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
