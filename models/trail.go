package models

import "time"

type Trail struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `json:"user_id"`
	Name      string `json:"name"`
	GeoJSON   string `json:"geojson"` // store as JSON string
	CreatedAt time.Time
	UpdatedAt time.Time
}
