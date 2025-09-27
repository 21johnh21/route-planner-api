package handlers

import (
	"net/http"
	"time"

	"github.com/21johnh21/route-planner-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetTrails returns all trails for the authenticated user
func GetTrails(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(float64) // JWT returns float64 for numeric IDs

	var trails []models.Trail
	if err := db.Where("user_id = ?", uint(userID)).Find(&trails).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trails"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trails": trails})
}

// CreateTrail stores a new trail for the authenticated user
func CreateTrail(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(float64)

	var input struct {
		Name    string `json:"name" binding:"required"`
		GeoJSON string `json:"geojson" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trail := models.Trail{
		UserID:    uint(userID),
		Name:      input.Name,
		GeoJSON:   input.GeoJSON,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&trail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save trail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trail": trail})
}
