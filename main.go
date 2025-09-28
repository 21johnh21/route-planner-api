package main

import (
	"log"
	"os"

	"time"

	"github.com/21johnh21/route-planner-api/config"
	"github.com/21johnh21/route-planner-api/routes"
	"github.com/gin-contrib/cors"
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
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.SetTrustedProxies([]string{"127.0.0.1"})

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
