package repository

import (
	"database/sql"
	"time"

	"officestonks/internal/models"
)

// ActivityRepo implements models.ActivityRepository
type ActivityRepo struct {
	db *sql.DB
}

// NewActivityRepo creates a new ActivityRepo
func NewActivityRepo(db *sql.DB) *ActivityRepo {
	return &ActivityRepo{db: db}
}

// LogActivity records a user activity event
func (r *ActivityRepo) LogActivity(userID int, username, action, details, ipAddress string, success bool, errorMsg string) error {
	query := `
		INSERT INTO user_activity (user_id, username, action, details, ip_address, success, error_message) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := RetryExec(r.db, query, userID, username, action, details, ipAddress, success, errorMsg)
	return err
}

// GetRecentActivity returns recent activity across all users
func (r *ActivityRepo) GetRecentActivity(limit int) ([]*models.UserActivity, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, user_id, username, action, details, ip_address, 
		       timestamp, success, error_message
		FROM user_activity
		ORDER BY timestamp DESC
		LIMIT ?
	`
	rows, err := RetryQuery(r.db, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.UserActivity
	for rows.Next() {
		var a models.UserActivity
		var details, errorMsg sql.NullString
		err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.Action, &details, 
			&a.IPAddress, &a.Timestamp, &a.Success, &errorMsg)
		if err != nil {
			return nil, err
		}
		if details.Valid {
			a.Details = details.String
		}
		if errorMsg.Valid {
			a.ErrorMsg = errorMsg.String
		}
		activities = append(activities, &a)
	}
	return activities, nil
}

// GetUserActivity returns recent activity for a specific user
func (r *ActivityRepo) GetUserActivity(userID int, limit int) ([]*models.UserActivity, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT id, user_id, username, action, details, ip_address,
		       timestamp, success, error_message
		FROM user_activity
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`
	rows, err := RetryQuery(r.db, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.UserActivity
	for rows.Next() {
		var a models.UserActivity
		var details, errorMsg sql.NullString
		err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.Action, &details,
			&a.IPAddress, &a.Timestamp, &a.Success, &errorMsg)
		if err != nil {
			return nil, err
		}
		if details.Valid {
			a.Details = details.String
		}
		if errorMsg.Valid {
			a.ErrorMsg = errorMsg.String
		}
		activities = append(activities, &a)
	}
	return activities, nil
}

// GetActivityByTimeRange returns activity within a time range
func (r *ActivityRepo) GetActivityByTimeRange(startTime, endTime time.Time) ([]*models.UserActivity, error) {
	query := `
		SELECT id, user_id, username, action, details, ip_address,
		       timestamp, success, error_message
		FROM user_activity
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
		LIMIT 1000
	`
	rows, err := RetryQuery(r.db, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.UserActivity
	for rows.Next() {
		var a models.UserActivity
		var details, errorMsg sql.NullString
		err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.Action, &details,
			&a.IPAddress, &a.Timestamp, &a.Success, &errorMsg)
		if err != nil {
			return nil, err
		}
		if details.Valid {
			a.Details = details.String
		}
		if errorMsg.Valid {
			a.ErrorMsg = errorMsg.String
		}
		activities = append(activities, &a)
	}
	return activities, nil
}