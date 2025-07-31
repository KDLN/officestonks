package models

import "time"

// Sector represents a market sector
type Sector struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Trend             float64   `json:"trend"`
	VolatilityModifier float64   `json:"volatility_modifier"`
	CreatedAt         time.Time `json:"created_at"`
}

// SectorRepository defines data access methods for sectors
type SectorRepository interface {
	GetAllSectors() ([]*Sector, error)
	GetSectorByID(id int) (*Sector, error)
	GetSectorByName(name string) (*Sector, error)
	UpdateSectorTrend(id int, trend float64) error
	CreateSector(name, description string) (*Sector, error)
	DeleteSector(id int) error
}