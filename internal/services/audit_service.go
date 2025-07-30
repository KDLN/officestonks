package services

import "officestonks/internal/models"

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
		return nil
	}
	return s.repo.LogEvent(userID, action, ip)
}

// GetRecentEvents returns recent audit events
func (s *AuditService) GetRecentEvents(limit int) ([]*models.AuditEvent, error) {
	if s.repo == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetRecentEvents(limit)
}
