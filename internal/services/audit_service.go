package services

import (
	"fmt"

	"officestonks/internal/models"
)

// AuditService provides audit logging operations
type AuditService struct {
	repo models.AuditLogRepository
}

// NewAuditService creates an AuditService
func NewAuditService(repo models.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

// LogEvent records an audit event
func (s *AuditService) LogEvent(userID int, action, ip string) error {
	if s.repo == nil {
		return fmt.Errorf("audit repository not configured")
	}
	return s.repo.LogEvent(userID, action, ip)
}

// GetRecentEvents returns recent audit events
func (s *AuditService) GetRecentEvents(limit int) ([]*models.AuditEvent, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("audit repository not configured")
	}
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetRecentEvents(limit)
}
