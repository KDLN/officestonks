package models

import "time"

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Action    string    `json:"action"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLogRepository defines persistence methods for audit events
type AuditLogRepository interface {
	LogEvent(userID int, action, ipAddress string) error
	GetRecentEvents(limit int) ([]*AuditEvent, error)
}
