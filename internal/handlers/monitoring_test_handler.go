package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// MonitoringTestHandler provides test endpoints for monitoring system
type MonitoringTestHandler struct {
	db *sql.DB
}

// NewMonitoringTestHandler creates a new monitoring test handler
func NewMonitoringTestHandler(db *sql.DB) *MonitoringTestHandler {
	return &MonitoringTestHandler{db: db}
}

// TestMonitoring checks if monitoring tables exist and have data
func (h *MonitoringTestHandler) TestMonitoring(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	result := map[string]interface{}{
		"timestamp": time.Now(),
		"checks":    []map[string]interface{}{},
	}

	// Check if user_sessions table exists
	var tableName string
	err := h.db.QueryRow("SHOW TABLES LIKE 'user_sessions'").Scan(&tableName)
	sessionsTableExists := err == nil
	result["checks"] = append(result["checks"].([]map[string]interface{}), map[string]interface{}{
		"name":   "user_sessions table exists",
		"status": sessionsTableExists,
		"error":  errorToString(err),
	})

	// Count sessions if table exists
	if sessionsTableExists {
		var count int
		err = h.db.QueryRow("SELECT COUNT(*) FROM user_sessions").Scan(&count)
		result["checks"] = append(result["checks"].([]map[string]interface{}), map[string]interface{}{
			"name":   "user_sessions count",
			"status": err == nil,
			"value":  count,
			"error":  errorToString(err),
		})

		// Count active sessions
		err = h.db.QueryRow("SELECT COUNT(*) FROM user_sessions WHERE is_active = TRUE").Scan(&count)
		result["checks"] = append(result["checks"].([]map[string]interface{}), map[string]interface{}{
			"name":   "active sessions count",
			"status": err == nil,
			"value":  count,
			"error":  errorToString(err),
		})
	}

	// Check if user_activity table exists
	err = h.db.QueryRow("SHOW TABLES LIKE 'user_activity'").Scan(&tableName)
	activityTableExists := err == nil
	result["checks"] = append(result["checks"].([]map[string]interface{}), map[string]interface{}{
		"name":   "user_activity table exists",
		"status": activityTableExists,
		"error":  errorToString(err),
	})

	// Count activities if table exists
	if activityTableExists {
		var count int
		err = h.db.QueryRow("SELECT COUNT(*) FROM user_activity").Scan(&count)
		result["checks"] = append(result["checks"].([]map[string]interface{}), map[string]interface{}{
			"name":   "user_activity count",
			"status": err == nil,
			"value":  count,
			"error":  errorToString(err),
		})
	}

	// Check if users table has new columns
	var columnCount int
	err = h.db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME IN ('last_login', 'login_count', 'total_trades')").Scan(&columnCount)
	result["checks"] = append(result["checks"].([]map[string]interface{}), map[string]interface{}{
		"name":   "users table has monitoring columns",
		"status": columnCount == 3,
		"value":  columnCount,
		"error":  errorToString(err),
	})

	// Test creating a session
	if sessionsTableExists {
		_, err = h.db.Exec("INSERT INTO user_sessions (user_id, ip_address, user_agent) VALUES (?, ?, ?)", 
			1, "127.0.0.1", "Test Agent")
		canInsert := err == nil
		result["checks"] = append(result["checks"].([]map[string]interface{}), map[string]interface{}{
			"name":   "can insert into user_sessions",
			"status": canInsert,
			"error":  errorToString(err),
		})

		if canInsert {
			// Clean up test session
			h.db.Exec("DELETE FROM user_sessions WHERE user_agent = 'Test Agent'")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}