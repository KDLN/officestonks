package repository

import (
	"database/sql"

	"officestonks/internal/models"
)

// SectorRepo implements models.SectorRepository backed by MySQL
type SectorRepo struct {
	db *sql.DB
}

// NewSectorRepo creates a new SectorRepo
func NewSectorRepo(db *sql.DB) *SectorRepo {
	return &SectorRepo{db: db}
}

// GetAllSectors returns all sectors
func (r *SectorRepo) GetAllSectors() ([]*models.Sector, error) {
	query := `
		SELECT id, name, description, trend, volatility_modifier, created_at
		FROM sectors
		ORDER BY name
	`
	rows, err := RetryQuery(r.db, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sectors []*models.Sector
	for rows.Next() {
		var s models.Sector
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Trend, &s.VolatilityModifier, &s.CreatedAt); err != nil {
			return nil, err
		}
		sectors = append(sectors, &s)
	}
	return sectors, nil
}

// GetSectorByID returns a sector by ID
func (r *SectorRepo) GetSectorByID(id int) (*models.Sector, error) {
	query := `
		SELECT id, name, description, trend, volatility_modifier, created_at
		FROM sectors
		WHERE id = ?
	`
	var s models.Sector
	err := RetryQueryRow(r.db, query, id).Scan(&s.ID, &s.Name, &s.Description, &s.Trend, &s.VolatilityModifier, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSectorByName returns a sector by name
func (r *SectorRepo) GetSectorByName(name string) (*models.Sector, error) {
	query := `
		SELECT id, name, description, trend, volatility_modifier, created_at
		FROM sectors
		WHERE name = ?
	`
	var s models.Sector
	err := RetryQueryRow(r.db, query, name).Scan(&s.ID, &s.Name, &s.Description, &s.Trend, &s.VolatilityModifier, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSectorTrend updates a sector's trend value
func (r *SectorRepo) UpdateSectorTrend(id int, trend float64) error {
	query := "UPDATE sectors SET trend = ? WHERE id = ?"
	_, err := RetryExec(r.db, query, trend, id)
	return err
}

// CreateSector creates a new sector
func (r *SectorRepo) CreateSector(name, description string) (*models.Sector, error) {
	query := `
		INSERT INTO sectors (name, description)
		VALUES (?, ?)
	`
	result, err := RetryExec(r.db, query, name, description)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetSectorByID(int(id))
}

// DeleteSector deletes a sector
func (r *SectorRepo) DeleteSector(id int) error {
	query := "DELETE FROM sectors WHERE id = ?"
	_, err := RetryExec(r.db, query, id)
	return err
}