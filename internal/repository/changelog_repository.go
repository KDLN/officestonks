package repository

import (
	"database/sql"
	"encoding/json"

	"officestonks/internal/models"
)

// ChangelogRepo implements models.ChangelogRepository backed by MySQL
type ChangelogRepo struct {
	db *sql.DB
}

// NewChangelogRepo creates a new ChangelogRepo
func NewChangelogRepo(db *sql.DB) *ChangelogRepo {
	return &ChangelogRepo{db: db}
}

// CreateEntry creates a new changelog entry
func (r *ChangelogRepo) CreateEntry(version, title, description string, changes []string, changeType models.ChangeType, isMajor bool, createdBy *int) (*models.ChangelogEntry, error) {
	// Convert changes slice to JSON
	changesJSON, err := json.Marshal(changes)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO changelog (version, title, description, changes, change_type, is_major, is_visible, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := RetryExec(r.db, query, version, title, description, changesJSON, string(changeType), isMajor, true, createdBy)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return the created entry
	return r.GetEntryByID(int(id))
}

// GetAllEntries returns all changelog entries with pagination
func (r *ChangelogRepo) GetAllEntries(limit, offset int) ([]*models.ChangelogEntry, error) {
	query := `
		SELECT id, version, title, description, changes, change_type, is_major, is_visible, created_at, created_by
		FROM changelog
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := RetryQuery(r.db, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChangelogEntries(rows)
}

// GetVisibleEntries returns only visible changelog entries with pagination
func (r *ChangelogRepo) GetVisibleEntries(limit, offset int) ([]*models.ChangelogEntry, error) {
	query := `
		SELECT id, version, title, description, changes, change_type, is_major, is_visible, created_at, created_by
		FROM changelog
		WHERE is_visible = true
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := RetryQuery(r.db, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChangelogEntries(rows)
}

// GetMajorEntries returns only major release entries
func (r *ChangelogRepo) GetMajorEntries(limit int) ([]*models.ChangelogEntry, error) {
	query := `
		SELECT id, version, title, description, changes, change_type, is_major, is_visible, created_at, created_by
		FROM changelog
		WHERE is_major = true AND is_visible = true
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := RetryQuery(r.db, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChangelogEntries(rows)
}

// GetEntryByID returns a specific changelog entry by ID
func (r *ChangelogRepo) GetEntryByID(id int) (*models.ChangelogEntry, error) {
	query := `
		SELECT id, version, title, description, changes, change_type, is_major, is_visible, created_at, created_by
		FROM changelog
		WHERE id = ?
	`
	
	row := RetryQueryRow(r.db, query, id)
	return r.scanSingleChangelogEntry(row)
}

// GetEntryByVersion returns a specific changelog entry by version
func (r *ChangelogRepo) GetEntryByVersion(version string) (*models.ChangelogEntry, error) {
	query := `
		SELECT id, version, title, description, changes, change_type, is_major, is_visible, created_at, created_by
		FROM changelog
		WHERE version = ?
	`
	
	row := RetryQueryRow(r.db, query, version)
	return r.scanSingleChangelogEntry(row)
}

// UpdateEntryVisibility updates the visibility of a changelog entry
func (r *ChangelogRepo) UpdateEntryVisibility(id int, isVisible bool) error {
	query := `UPDATE changelog SET is_visible = ? WHERE id = ?`
	_, err := RetryExec(r.db, query, isVisible, id)
	return err
}

// DeleteEntry deletes a changelog entry
func (r *ChangelogRepo) DeleteEntry(id int) error {
	query := `DELETE FROM changelog WHERE id = ?`
	_, err := RetryExec(r.db, query, id)
	return err
}

// scanChangelogEntries scans multiple changelog entries from rows
func (r *ChangelogRepo) scanChangelogEntries(rows *sql.Rows) ([]*models.ChangelogEntry, error) {
	var entries []*models.ChangelogEntry
	
	for rows.Next() {
		var entry models.ChangelogEntry
		var changesJSON []byte
		var changeTypeStr string
		
		err := rows.Scan(
			&entry.ID,
			&entry.Version,
			&entry.Title,
			&entry.Description,
			&changesJSON,
			&changeTypeStr,
			&entry.IsMajor,
			&entry.IsVisible,
			&entry.CreatedAt,
			&entry.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON changes
		err = json.Unmarshal(changesJSON, &entry.Changes)
		if err != nil {
			return nil, err
		}

		// Convert string to ChangeType
		entry.ChangeType = models.ChangeType(changeTypeStr)

		entries = append(entries, &entry)
	}

	return entries, nil
}

// scanSingleChangelogEntry scans a single changelog entry from a row
func (r *ChangelogRepo) scanSingleChangelogEntry(row *sql.Row) (*models.ChangelogEntry, error) {
	var entry models.ChangelogEntry
	var changesJSON []byte
	var changeTypeStr string
	
	err := row.Scan(
		&entry.ID,
		&entry.Version,
		&entry.Title,
		&entry.Description,
		&changesJSON,
		&changeTypeStr,
		&entry.IsMajor,
		&entry.IsVisible,
		&entry.CreatedAt,
		&entry.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Parse JSON changes
	err = json.Unmarshal(changesJSON, &entry.Changes)
	if err != nil {
		return nil, err
	}

	// Convert string to ChangeType
	entry.ChangeType = models.ChangeType(changeTypeStr)

	return &entry, nil
}