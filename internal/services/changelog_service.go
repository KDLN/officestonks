package services

import (
	"officestonks/internal/models"
)

// ChangelogService handles changelog operations
type ChangelogService struct {
	changelogRepo models.ChangelogRepository
}

// NewChangelogService creates a new changelog service
func NewChangelogService(changelogRepo models.ChangelogRepository) *ChangelogService {
	return &ChangelogService{
		changelogRepo: changelogRepo,
	}
}

// CreateEntry creates a new changelog entry
func (s *ChangelogService) CreateEntry(version, title, description string, changes []string, changeType models.ChangeType, isMajor bool, createdBy *int) (*models.ChangelogEntry, error) {
	return s.changelogRepo.CreateEntry(version, title, description, changes, changeType, isMajor, createdBy)
}

// GetVisibleEntries returns visible changelog entries for public display
func (s *ChangelogService) GetVisibleEntries(limit, offset int) ([]*models.ChangelogEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}
	return s.changelogRepo.GetVisibleEntries(limit, offset)
}

// GetMajorEntries returns major release entries
func (s *ChangelogService) GetMajorEntries(limit int) ([]*models.ChangelogEntry, error) {
	if limit <= 0 || limit > 50 {
		limit = 10 // Default limit for major releases
	}
	return s.changelogRepo.GetMajorEntries(limit)
}

// GetAllEntries returns all changelog entries (admin only)
func (s *ChangelogService) GetAllEntries(limit, offset int) ([]*models.ChangelogEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}
	return s.changelogRepo.GetAllEntries(limit, offset)
}

// GetEntryByVersion returns a specific entry by version
func (s *ChangelogService) GetEntryByVersion(version string) (*models.ChangelogEntry, error) {
	return s.changelogRepo.GetEntryByVersion(version)
}

// UpdateEntryVisibility updates the visibility of an entry (admin only)
func (s *ChangelogService) UpdateEntryVisibility(id int, isVisible bool) error {
	return s.changelogRepo.UpdateEntryVisibility(id, isVisible)
}

// DeleteEntry deletes a changelog entry (admin only)
func (s *ChangelogService) DeleteEntry(id int) error {
	return s.changelogRepo.DeleteEntry(id)
}