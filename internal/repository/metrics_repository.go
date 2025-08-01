package repository

import (
	"database/sql"
	"time"

	"officestonks/internal/models"
)

// MetricsRepo implements models.MetricsRepository
type MetricsRepo struct {
	db *sql.DB
}

// NewMetricsRepo creates a new MetricsRepo
func NewMetricsRepo(db *sql.DB) *MetricsRepo {
	return &MetricsRepo{db: db}
}

// GetSystemMetrics returns current system performance metrics
func (r *MetricsRepo) GetSystemMetrics() (*models.SystemMetrics, error) {
	activeUsers, _ := r.GetActiveUserCount()
	tradesPerHour, _ := r.GetTradesPerHour()
	errorRate, _ := r.GetErrorRate()

	totalSessions, err := r.getTotalSessions()
	if err != nil {
		totalSessions = 0
	}

	activeSessions, err := r.getActiveSessions()
	if err != nil {
		activeSessions = 0
	}

	websocketConns, err := r.getActiveWebSocketConnections()
	if err != nil {
		websocketConns = 0
	}

	dbHealth := r.getDatabaseHealth()
	avgResponseTime, _ := r.getAverageResponseTime()

	return &models.SystemMetrics{
		Timestamp:        time.Now(),
		ActiveUsers:      activeUsers,
		TotalSessions:    totalSessions,
		ActiveSessions:   activeSessions,
		TradesPerHour:    tradesPerHour,
		WebSocketConns:   websocketConns,
		DatabaseHealth:   dbHealth,
		ErrorRate:        errorRate,
		AvgResponseTime:  avgResponseTime,
	}, nil
}

// GetActiveUserCount returns the number of currently active users
func (r *MetricsRepo) GetActiveUserCount() (int, error) {
	query := `SELECT COUNT(DISTINCT user_id) FROM user_sessions WHERE is_active = TRUE`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTradesPerHour returns the number of trades in the last hour
func (r *MetricsRepo) GetTradesPerHour() (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM transactions 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
	`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetErrorRate returns the error rate as a percentage in the last hour
func (r *MetricsRepo) GetErrorRate() (float64, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN success = FALSE THEN 1 END) as errors,
			COUNT(*) as total
		FROM user_activity 
		WHERE timestamp >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
	`
	var errors, total int
	err := r.db.QueryRow(query).Scan(&errors, &total)
	if err != nil || total == 0 {
		return 0.0, err
	}
	return float64(errors) / float64(total) * 100, nil
}

// Helper methods for GetSystemMetrics

func (r *MetricsRepo) getTotalSessions() (int, error) {
	query := `SELECT COUNT(*) FROM user_sessions WHERE DATE(login_time) = CURDATE()`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

func (r *MetricsRepo) getActiveSessions() (int, error) {
	query := `SELECT COUNT(*) FROM user_sessions WHERE is_active = TRUE`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

func (r *MetricsRepo) getActiveWebSocketConnections() (int, error) {
	// Try to get from websocket_connections table if it exists
	query := `SELECT COUNT(*) FROM websocket_connections WHERE is_active = TRUE`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		// Fallback to active sessions as an approximation
		return r.getActiveSessions()
	}
	return count, nil
}

func (r *MetricsRepo) getDatabaseHealth() string {
	// Simple health check - try to query the database
	var result int
	err := r.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return "down"
	}
	
	// Check if response time is reasonable (under 100ms)
	start := time.Now()
	err = r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&result)
	duration := time.Since(start)
	
	if err != nil {
		return "degraded"
	}
	
	if duration > 100*time.Millisecond {
		return "degraded"
	}
	
	return "healthy"
}

func (r *MetricsRepo) getAverageResponseTime() (float64, error) {
	query := `
		SELECT AVG(COALESCE(response_time_ms, 0)) 
		FROM user_activity 
		WHERE timestamp >= DATE_SUB(NOW(), INTERVAL 1 HOUR) 
		AND response_time_ms IS NOT NULL
	`
	var avgTime sql.NullFloat64
	err := r.db.QueryRow(query).Scan(&avgTime)
	if err != nil || !avgTime.Valid {
		return 0.0, err
	}
	return avgTime.Float64, nil
}

// RecordSystemMetrics stores current metrics in the database for historical tracking
func (r *MetricsRepo) RecordSystemMetrics() error {
	metrics, err := r.GetSystemMetrics()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO system_metrics 
		(active_users, total_sessions, active_sessions, trades_per_hour, 
		 websocket_connections, database_health, error_rate, avg_response_time_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = RetryExec(r.db, query, 
		metrics.ActiveUsers, metrics.TotalSessions, metrics.ActiveSessions,
		metrics.TradesPerHour, metrics.WebSocketConns, metrics.DatabaseHealth,
		metrics.ErrorRate, metrics.AvgResponseTime)
	
	return err
}