package models

import (
	"time"
)

type Tile struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Zoom             int       `json:"zoom"`
	X                int       `json:"x"`
	Y                int       `json:"y"`
	TrailGeoJSON     string    `json:"trail_geojson"`     // Serialized GeoJSON string
	TrailheadGeoJSON string    `json:"trailhead_geojson"` // Serialized GeoJSON string
	FetchedAt        time.Time `json:"fetched_at"`
}
