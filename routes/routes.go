package routes

import (
	"log"
	"time"

	"github.com/21johnh21/route-planner-api/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Simple request logging middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		log.Printf("[%d] %s %s %s", status, c.Request.Method, c.Request.URL.Path, latency)
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

		trail := api.Group("/trails")
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
