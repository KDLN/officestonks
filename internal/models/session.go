package models

import "time"

// UserSession represents a user login session
type UserSession struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Username    string    `json:"username,omitempty"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	LoginTime   time.Time `json:"login_time"`
	LogoutTime  *time.Time `json:"logout_time,omitempty"`
	IsActive    bool      `json:"is_active"`
	TradesCount int       `json:"trades_count"`
	LastActivity time.Time `json:"last_activity"`
}

// SystemMetrics represents real-time system performance data
type SystemMetrics struct {
	Timestamp        time.Time `json:"timestamp"`
	ActiveUsers      int       `json:"active_users"`
	TotalSessions    int       `json:"total_sessions"`
	ActiveSessions   int       `json:"active_sessions"`
	TradesPerHour    int       `json:"trades_per_hour"`
	WebSocketConns   int       `json:"websocket_connections"`
	DatabaseHealth   string    `json:"database_health"`
	ErrorRate        float64   `json:"error_rate"`
	AvgResponseTime  float64   `json:"avg_response_time_ms"`
}

// UserActivity represents detailed user activity for monitoring
type UserActivity struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Username   string    `json:"username,omitempty"`
	Action     string    `json:"action"`
	Details    string    `json:"details,omitempty"`
	IPAddress  string    `json:"ip_address"`
	Timestamp  time.Time `json:"timestamp"`
	Success    bool      `json:"success"`
	ErrorMsg   string    `json:"error_message,omitempty"`
}

// SessionRepository defines methods for session data access
type SessionRepository interface {
	CreateSession(userID int, ipAddress, userAgent string) (*UserSession, error)
	FindOrCreateActiveSession(userID int, ipAddress, userAgent string) (*UserSession, error)
	UpdateSessionActivity(sessionID int) error
	EndSession(sessionID int) error
	GetActiveSessions() ([]*UserSession, error)
	GetUserSessions(userID int, limit int) ([]*UserSession, error)
	IncrementTradeCount(sessionID int) error
	CleanupExpiredSessions() error
}

// ActivityRepository defines methods for user activity tracking
type ActivityRepository interface {
	LogActivity(userID int, username, action, details, ipAddress string, success bool, errorMsg string) error
	GetRecentActivity(limit int) ([]*UserActivity, error)
	GetUserActivity(userID int, limit int) ([]*UserActivity, error)
	GetActivityByTimeRange(startTime, endTime time.Time) ([]*UserActivity, error)
}

// MetricsRepository defines methods for system metrics
type MetricsRepository interface {
	GetSystemMetrics() (*SystemMetrics, error)
	GetActiveUserCount() (int, error)
	GetTradesPerHour() (int, error)
	GetErrorRate() (float64, error)
}