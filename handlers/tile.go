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

	// Simplified Overpass queries - request less data
	trailQuery := fmt.Sprintf(`
		[out:json][timeout:25];
		(
			way["highway"~"path|footway|cycleway|pedestrian|track|steps|bridleway"](%f,%f,%f,%f);
			relation["route"~"hiking|bicycle|foot"](%f,%f,%f,%f);
		);
		out body; >; out skel qt;
	`, latS, lonW, latN, lonE, latS, lonW, latN, lonE)

	trailheadQuery := fmt.Sprintf(`
		[out:json][timeout:25];
		(
			node["information"="trailhead"](%f,%f,%f,%f);
			node["amenity"="parking"]["hiking"="yes"](%f,%f,%f,%f);
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

	// Convert to optimized GeoJSON
	var trailData, trailheadData map[string]interface{}
	json.Unmarshal(trailBody, &trailData)
	json.Unmarshal(trailheadBody, &trailheadData)

	// Simplify and optimize the data
	simplifiedTrails := simplifyOverpassData(trailData, z)
	simplifiedTrailheads := simplifyOverpassData(trailheadData, z)

	trailGeoJSONBytes, err := json.Marshal(simplifiedTrails)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse trail data"})
		return
	}

	trailheadGeoJSONBytes, err := json.Marshal(simplifiedTrailheads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse trailhead data"})
		return
	}

	tile = models.Tile{
		Zoom:             z,
		X:                x,
		Y:                y,
		TrailGeoJSON:     string(trailGeoJSONBytes),
		TrailheadGeoJSON: string(trailheadGeoJSONBytes),
		FetchedAt:        time.Now(),
	}

	db.Save(&tile)

	c.JSON(http.StatusOK, gin.H{
		"trail_geojson":     tile.TrailGeoJSON,
		"trailhead_geojson": tile.TrailheadGeoJSON,
		"cached":            false,
	})
}

// simplifyOverpassData reduces data size by removing unnecessary fields and simplifying coordinates
func simplifyOverpassData(data map[string]interface{}, zoom int) map[string]interface{} {
	if data == nil {
		return map[string]interface{}{"elements": []interface{}{}}
	}

	elements, ok := data["elements"].([]interface{})
	if !ok {
		return data
	}

	// Determine coordinate precision based on zoom level
	// Higher zoom = more precision needed
	precision := getPrecisionForZoom(zoom)

	simplified := make([]interface{}, 0, len(elements))

	for _, e := range elements {
		elem, ok := e.(map[string]interface{})
		if !ok {
			continue
		}

		// Create a new simplified element with only essential fields
		newElem := map[string]interface{}{
			"type": elem["type"],
			"id":   elem["id"],
		}

		// Round coordinates to appropriate precision
		if lat, ok := elem["lat"].(float64); ok {
			newElem["lat"] = roundToPrecision(lat, precision)
		}
		if lon, ok := elem["lon"].(float64); ok {
			newElem["lon"] = roundToPrecision(lon, precision)
		}

		// Keep only essential tags
		if tags, ok := elem["tags"].(map[string]interface{}); ok {
			essentialTags := extractEssentialTags(tags)
			if len(essentialTags) > 0 {
				newElem["tags"] = essentialTags
			}
		}

		// Keep node references for ways (but they're already minimal)
		if nodes, ok := elem["nodes"].([]interface{}); ok {
			newElem["nodes"] = nodes
		}

		// Keep members for relations
		if members, ok := elem["members"].([]interface{}); ok {
			newElem["members"] = members
		}

		simplified = append(simplified, newElem)
	}

	return map[string]interface{}{
		"version":   data["version"],
		"generator": data["generator"],
		"elements":  simplified,
	}
}

// getPrecisionForZoom returns decimal places based on zoom level
func getPrecisionForZoom(zoom int) int {
	// At zoom 10: ~10m precision (5 decimal places)
	// At zoom 15: ~1m precision (6 decimal places)
	// At zoom 18: ~0.1m precision (7 decimal places)
	if zoom <= 10 {
		return 4
	} else if zoom <= 13 {
		return 5
	} else if zoom <= 16 {
		return 6
	}
	return 7
}

// roundToPrecision rounds a float to n decimal places
func roundToPrecision(val float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(val*multiplier) / multiplier
}

// extractEssentialTags keeps only important tags for UI display
func extractEssentialTags(tags map[string]interface{}) map[string]interface{} {
	essential := map[string]interface{}{}

	// List of tags that are useful for the UI
	keepTags := []string{
		"name", "highway", "surface", "access",
		"route", "information", "amenity", "parking",
		"hiking", "bicycle", "foot", "difficulty",
		"description", "leisure",
	}

	for _, key := range keepTags {
		if val, exists := tags[key]; exists {
			essential[key] = val
		}
	}

	return essential
}

// Helper functions
func exp(x float64) float64  { return math.Exp(x) }
func atan(x float64) float64 { return math.Atan(x) }
