package models

import "time"

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeTypeFeature    ChangeType = "feature"    // New features
	ChangeTypeImprovement ChangeType = "improvement" // Enhancements
	ChangeTypeBugfix     ChangeType = "bugfix"     // Bug fixes
	ChangeTypeBreaking   ChangeType = "breaking"   // Breaking changes
)

// ChangelogEntry represents a single changelog entry
type ChangelogEntry struct {
	ID          int         `json:"id"`
	Version     string      `json:"version"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Changes     []string    `json:"changes"`
	ChangeType  ChangeType  `json:"change_type"`
	IsMajor     bool        `json:"is_major"`
	IsVisible   bool        `json:"is_visible"`
	CreatedAt   time.Time   `json:"created_at"`
	CreatedBy   *int        `json:"created_by,omitempty"`
}

// ChangelogRepository defines data access methods for changelog
type ChangelogRepository interface {
	CreateEntry(version, title, description string, changes []string, changeType ChangeType, isMajor bool, createdBy *int) (*ChangelogEntry, error)
	GetAllEntries(limit, offset int) ([]*ChangelogEntry, error)
	GetVisibleEntries(limit, offset int) ([]*ChangelogEntry, error)
	GetMajorEntries(limit int) ([]*ChangelogEntry, error)
	GetEntryByID(id int) (*ChangelogEntry, error)
	GetEntryByVersion(version string) (*ChangelogEntry, error)
	UpdateEntryVisibility(id int, isVisible bool) error
	DeleteEntry(id int) error
}