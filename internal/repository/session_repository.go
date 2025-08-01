package repository

import (
	"database/sql"
	"time"

	"officestonks/internal/models"
)

// SessionRepo implements models.SessionRepository
type SessionRepo struct {
	db *sql.DB
}

// NewSessionRepo creates a new SessionRepo
func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// CreateSession creates a new user session
func (r *SessionRepo) CreateSession(userID int, ipAddress, userAgent string) (*models.UserSession, error) {
	query := `INSERT INTO user_sessions (user_id, ip_address, user_agent) VALUES (?, ?, ?)`
	result, err := RetryExec(r.db, query, userID, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	sessionID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update user's last_login and login_count
	updateUserQuery := `UPDATE users SET last_login = CURRENT_TIMESTAMP, login_count = login_count + 1 WHERE id = ?`
	_, err = RetryExec(r.db, updateUserQuery, userID)
	if err != nil {
		// Log error but don't fail the session creation
	}

	return &models.UserSession{
		ID:           int(sessionID),
		UserID:       userID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LoginTime:    time.Now(),
		IsActive:     true,
		TradesCount:  0,
		LastActivity: time.Now(),
	}, nil
}

// UpdateSessionActivity updates the last activity time for a session
func (r *SessionRepo) UpdateSessionActivity(sessionID int) error {
	query := `UPDATE user_sessions SET last_activity = CURRENT_TIMESTAMP WHERE id = ? AND is_active = TRUE`
	_, err := RetryExec(r.db, query, sessionID)
	return err
}

// EndSession marks a session as inactive
func (r *SessionRepo) EndSession(sessionID int) error {
	query := `UPDATE user_sessions SET is_active = FALSE, logout_time = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := RetryExec(r.db, query, sessionID)
	return err
}

// GetActiveSessions returns all currently active sessions
func (r *SessionRepo) GetActiveSessions() ([]*models.UserSession, error) {
	query := `
		SELECT s.id, s.user_id, u.username, s.ip_address, s.user_agent, 
		       s.login_time, s.logout_time, s.is_active, s.trades_count, s.last_activity
		FROM user_sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.is_active = TRUE
		ORDER BY s.last_activity DESC
	`
	rows, err := RetryQuery(r.db, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.UserSession
	for rows.Next() {
		var s models.UserSession
		var logoutTime sql.NullTime
		err := rows.Scan(&s.ID, &s.UserID, &s.Username, &s.IPAddress, &s.UserAgent,
			&s.LoginTime, &logoutTime, &s.IsActive, &s.TradesCount, &s.LastActivity)
		if err != nil {
			return nil, err
		}
		if logoutTime.Valid {
			s.LogoutTime = &logoutTime.Time
		}
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

// GetUserSessions returns recent sessions for a specific user
func (r *SessionRepo) GetUserSessions(userID int, limit int) ([]*models.UserSession, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT s.id, s.user_id, u.username, s.ip_address, s.user_agent,
		       s.login_time, s.logout_time, s.is_active, s.trades_count, s.last_activity
		FROM user_sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.user_id = ?
		ORDER BY s.login_time DESC
		LIMIT ?
	`
	rows, err := RetryQuery(r.db, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.UserSession
	for rows.Next() {
		var s models.UserSession
		var logoutTime sql.NullTime
		err := rows.Scan(&s.ID, &s.UserID, &s.Username, &s.IPAddress, &s.UserAgent,
			&s.LoginTime, &logoutTime, &s.IsActive, &s.TradesCount, &s.LastActivity)
		if err != nil {
			return nil, err
		}
		if logoutTime.Valid {
			s.LogoutTime = &logoutTime.Time
		}
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

// IncrementTradeCount increments the trade counter for a session
func (r *SessionRepo) IncrementTradeCount(sessionID int) error {
	query := `UPDATE user_sessions SET trades_count = trades_count + 1, last_activity = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := RetryExec(r.db, query, sessionID)
	return err
}

// CleanupExpiredSessions marks sessions as inactive if they haven't been active for more than 24 hours
func (r *SessionRepo) CleanupExpiredSessions() error {
	query := `
		UPDATE user_sessions 
		SET is_active = FALSE, logout_time = CURRENT_TIMESTAMP 
		WHERE is_active = TRUE AND last_activity < DATE_SUB(NOW(), INTERVAL 24 HOUR)
	`
	_, err := RetryExec(r.db, query)
	return err
}