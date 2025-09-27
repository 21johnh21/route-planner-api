package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/21johnh21/route-planner-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const CacheTTL = 24 * time.Hour

// GetTile handles GET /tiles/:z/:x/:y and returns cached or fresh tile GeoJSON
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

	// Convert tile coordinates to approximate bounding box (lat/lon)
	n := 1 << z
	lonW := float64(x)/float64(n)*360 - 180
	lonE := float64(x+1)/float64(n)*360 - 180
	latN := (180 / math.Pi) * (2*atan(exp(math.Pi*(1-2*float64(y)/float64(n)))) - math.Pi/2)
	latS := (180 / math.Pi) * (2*atan(exp(math.Pi*(1-2*float64(y+1)/float64(n)))) - math.Pi/2)

	// Overpass queries
	trailQuery := fmt.Sprintf(`
		[out:json][timeout:25];
		(
			way["highway"~"path|footway|cycleway|pedestrian|track|steps|bridleway"](%f,%f,%f,%f);
			relation["route"~"hiking|bicycle|foot"](%f,%f,%f,%f);
			relation["leisure"="park"](%f,%f,%f,%f);
		);
		out body; >; out skel qt;
	`, latS, lonW, latN, lonE, latS, lonW, latN, lonE, latS, lonW, latN, lonE)

	trailheadQuery := fmt.Sprintf(`
		[out:json][timeout:25];
		(
			node["information"="trailhead"]["informal"!="yes"]["parking"!="" ](%f,%f,%f,%f);
			node["amenity"="parking"]["access"!="private"]["hiking"="yes"](%f,%f,%f,%f);
		);
		out geom;
	`, latS, lonW, latN, lonE, latS, lonW, latN, lonE)

	// Fetch data
	trailRes, err := http.Get("https://overpass-api.de/api/interpreter?data=" + url.QueryEscape(trailQuery))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trails"})
		return
	}
	defer trailRes.Body.Close()
	trailBody, _ := io.ReadAll(trailRes.Body)

	trailheadRes, err := http.Get("https://overpass-api.de/api/interpreter?data=" + url.QueryEscape(trailheadQuery))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trailheads"})
		return
	}
	defer trailheadRes.Body.Close()
	trailheadBody, _ := io.ReadAll(trailheadRes.Body)

	// Convert to JSON strings to store
	var trailData, trailheadData interface{}
	json.Unmarshal(trailBody, &trailData)
	json.Unmarshal(trailheadBody, &trailheadData)

	trailGeoJSON, _ := json.Marshal(trailData)
	trailheadGeoJSON, _ := json.Marshal(trailheadData)

	tile = models.Tile{
		Zoom:             z,
		X:                x,
		Y:                y,
		TrailGeoJSON:     string(trailGeoJSON),
		TrailheadGeoJSON: string(trailheadGeoJSON),
		FetchedAt:        time.Now(),
	}

	db.Save(&tile)

	c.JSON(http.StatusOK, gin.H{
		"trail_geojson":     tile.TrailGeoJSON,
		"trailhead_geojson": tile.TrailheadGeoJSON,
		"cached":            false,
	})
}

// Helper functions
func exp(x float64) float64  { return math.Exp(x) }
func atan(x float64) float64 { return math.Atan(x) }
