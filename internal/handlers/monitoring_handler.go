package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"officestonks/internal/models"
	"officestonks/internal/services"
)

// MonitoringHandler provides admin monitoring endpoints
type MonitoringHandler struct {
	monitoringService *services.MonitoringService
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(monitoringService *services.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
	}
}

// Helper function to send JSON error responses
func (h *MonitoringHandler) sendJSONError(w http.ResponseWriter, message string, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"error": message,
	}
	if err != nil {
		response["details"] = err.Error()
	}
	json.NewEncoder(w).Encode(response)
}

// GetSystemMetrics returns real-time system performance metrics
func (h *MonitoringHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	metrics, err := h.monitoringService.GetSystemMetrics()
	if err != nil {
		h.sendJSONError(w, "Failed to fetch system metrics", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetActiveSessions returns all currently active user sessions
func (h *MonitoringHandler) GetActiveSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessions, err := h.monitoringService.GetActiveSessions()
	if err != nil {
		h.sendJSONError(w, "Failed to fetch active sessions", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// GetRecentActivity returns recent user activity across the system
func (h *MonitoringHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		if l > 500 {
			l = 500 // Cap at 500 for performance
		}
		limit = l
	}

	activities, err := h.monitoringService.GetRecentActivity(limit)
	if err != nil {
		h.sendJSONError(w, "Failed to fetch recent activity", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
	})
}

// GetUserActivity returns activity for a specific user
func (h *MonitoringHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		h.sendJSONError(w, "user_id parameter is required", http.StatusBadRequest, nil)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.sendJSONError(w, "Invalid user_id", http.StatusBadRequest, err)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		if l > 100 {
			l = 100
		}
		limit = l
	}

	activities, err := h.monitoringService.GetUserActivity(userID, limit)
	if err != nil {
		h.sendJSONError(w, "Failed to fetch user activity", http.StatusInternalServerError, err)
		return
	}

	sessions, err := h.monitoringService.GetUserSessions(userID, 5)
	if err != nil {
		sessions = []*models.UserSession{} // Don't fail if sessions can't be fetched
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":         userID,
		"activities":      activities,
		"recent_sessions": sessions,
		"activity_count":  len(activities),
	})
}

// GetUserSessions returns recent sessions for a specific user
func (h *MonitoringHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		h.sendJSONError(w, "user_id parameter is required", http.StatusBadRequest, nil)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.sendJSONError(w, "Invalid user_id", http.StatusBadRequest, err)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		if l > 50 {
			l = 50
		}
		limit = l
	}

	sessions, err := h.monitoringService.GetUserSessions(userID, limit)
	if err != nil {
		h.sendJSONError(w, "Failed to fetch user sessions", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  userID,
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// GetActivityByTimeRange returns activity within a specific time range
func (h *MonitoringHandler) GetActivityByTimeRange(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		h.sendJSONError(w, "start_time and end_time parameters are required", http.StatusBadRequest, nil)
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		h.sendJSONError(w, "Invalid start_time format. Use RFC3339", http.StatusBadRequest, err)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		h.sendJSONError(w, "Invalid end_time format. Use RFC3339", http.StatusBadRequest, err)
		return
	}

	// Limit time range to prevent excessive queries
	if endTime.Sub(startTime) > 7*24*time.Hour {
		h.sendJSONError(w, "Time range cannot exceed 7 days", http.StatusBadRequest, nil)
		return
	}

	activities, err := h.monitoringService.GetActivityByTimeRange(startTime, endTime)
	if err != nil {
		h.sendJSONError(w, "Failed to fetch activity by time range", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"start_time":  startTime,
		"end_time":    endTime,
		"activities":  activities,
		"count":       len(activities),
	})
}

// GetMonitoringDashboard returns a comprehensive dashboard summary
func (h *MonitoringHandler) GetMonitoringDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get system metrics
	metrics, err := h.monitoringService.GetSystemMetrics()
	if err != nil {
		h.sendJSONError(w, "Failed to fetch system metrics", http.StatusInternalServerError, err)
		return
	}

	// Get active sessions
	sessions, err := h.monitoringService.GetActiveSessions()
	if err != nil {
		sessions = []*models.UserSession{} // Don't fail completely
	}

	// Get recent activity (last 25 entries)
	activities, err := h.monitoringService.GetRecentActivity(25)
	if err != nil {
		activities = []*models.UserActivity{} // Don't fail completely
	}

	// Calculate some additional stats
	now := time.Now()
	recentLogins := 0
	recentTrades := 0
	recentErrors := 0

	for _, activity := range activities {
		if now.Sub(activity.Timestamp) <= time.Hour {
			switch activity.Action {
			case "login":
				recentLogins++
			case "trade":
				recentTrades++
			}
			if !activity.Success {
				recentErrors++
			}
		}
	}

	dashboard := map[string]interface{}{
		"timestamp":      now,
		"system_metrics": metrics,
		"active_sessions": map[string]interface{}{
			"sessions": sessions,
			"count":    len(sessions),
		},
		"recent_activity": map[string]interface{}{
			"activities": activities,
			"count":      len(activities),
		},
		"hourly_stats": map[string]interface{}{
			"logins": recentLogins,
			"trades": recentTrades,
			"errors": recentErrors,
		},
		"health_status": map[string]interface{}{
			"database":         metrics.DatabaseHealth,
			"error_rate":       metrics.ErrorRate,
			"avg_response_ms":  metrics.AvgResponseTime,
			"websocket_conns":  metrics.WebSocketConns,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}