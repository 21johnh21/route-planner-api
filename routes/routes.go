package routes

import (
	"log"
	"time"

	"github.com/21johnh21/route-planner-api/handlers"
	"github.com/21johnh21/route-planner-api/middleware" // <--- import middleware
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Simple request logging middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		// Only log summary, truncate query or omit headers/body
		log.Printf("[%d] %s %s (%s)", status, method, path, latency)
	})

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/signup", handlers.SignUp)
			auth.POST("/login", handlers.Login)
		}

		// Apply AuthMiddleware here
		trail := api.Group("/trails")
		trail.Use(middleware.AuthMiddleware())
		{
			trail.GET("/", handlers.GetTrails)
			trail.POST("/", handlers.CreateTrail)
		}

		tile := api.Group("/tiles")
		{
			tile.GET("/:z/:x/:y", handlers.GetTile)
		}
	}
}
