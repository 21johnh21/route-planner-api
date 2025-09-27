package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"route-planner-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const CacheTTL = 24 * time.Hour

// GetTile handles GET /tile/:z/:x/:y and returns cached or fresh tile GeoJSON
func GetTile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	zStr := c.Param("z")
	xStr := c.Param("x")
	yStr := c.Param("y")

	z, err := strconv.Atoi(zStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zoom"})
		return
	}
	x, err := strconv.Atoi(xStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid x"})
		return
	}
	y, err := strconv.Atoi(yStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid y"})
		return
	}

	var tile models.Tile
	err = db.Where("zoom = ? AND x = ? AND y = ?", z, x, y).First(&tile).Error
	if err == nil && time.Since(tile.FetchedAt) < CacheTTL {
		c.JSON(http.StatusOK, gin.H{
			"trail_geojson":     tile.TrailGeoJSON,
			"trailhead_geojson": tile.TrailheadGeoJSON,
			"cached":            true,
		})
		return
	}

	// Placeholder: Fetch from Overpass API (implement real fetch later)
	trailGeoJSON := fmt.Sprintf(`{"type":"FeatureCollection","features":[]}`)
	trailheadGeoJSON := fmt.Sprintf(`{"type":"FeatureCollection","features":[]}`)

	tile = models.Tile{
		Zoom:             z,
		X:                x,
		Y:                y,
		TrailGeoJSON:     trailGeoJSON,
		TrailheadGeoJSON: trailheadGeoJSON,
		FetchedAt:        time.Now(),
	}

	db.Save(&tile)

	c.JSON(http.StatusOK, gin.H{
		"trail_geojson":     tile.TrailGeoJSON,
		"trailhead_geojson": tile.TrailheadGeoJSON,
		"cached":            false,
	})
}
