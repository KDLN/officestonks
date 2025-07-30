package repository

import (
	"database/sql"

	"officestonks/internal/models"
)

// AuditRepo implements models.AuditLogRepository
type AuditRepo struct {
	db *sql.DB
}

// NewAuditRepo creates a new AuditRepo
func NewAuditRepo(db *sql.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

// LogEvent inserts an audit log entry
func (r *AuditRepo) LogEvent(userID int, action, ipAddress string) error {
	query := `INSERT INTO audit_logs (user_id, action, ip_address) VALUES (?, ?, ?)`
	_, err := RetryExec(r.db, query, userID, action, ipAddress)
	return err
}

// GetRecentEvents returns the most recent audit events
func (r *AuditRepo) GetRecentEvents(limit int) ([]*models.AuditEvent, error) {
	query := `SELECT id, user_id, action, ip_address, created_at FROM audit_logs ORDER BY created_at DESC LIMIT ?`
	rows, err := RetryQuery(r.db, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.AuditEvent
	for rows.Next() {
		var evt models.AuditEvent
		if err := rows.Scan(&evt.ID, &evt.UserID, &evt.Action, &evt.IPAddress, &evt.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &evt)
	}
	return events, nil
}
