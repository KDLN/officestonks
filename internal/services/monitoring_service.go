package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"officestonks/internal/models"
	"officestonks/internal/repository"
)

// MonitoringService provides comprehensive system monitoring capabilities
type MonitoringService struct {
	sessionRepo  models.SessionRepository
	activityRepo models.ActivityRepository
	metricsRepo  models.MetricsRepository
	
	// In-memory tracking for real-time metrics
	activeConnections sync.Map // map[string]*ConnectionInfo
	requestMetrics    *RequestMetrics
	mu                sync.RWMutex
}

// ConnectionInfo tracks individual WebSocket connections
type ConnectionInfo struct {
	UserID      int
	Username    string
	ConnectedAt time.Time
	IPAddress   string
}

// RequestMetrics tracks API request performance
type RequestMetrics struct {
	TotalRequests   int64
	FailedRequests  int64
	TotalTime       time.Duration
	LastHourReqs    []time.Time
	LastHourErrors  []time.Time
	mu              sync.RWMutex
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(sessionRepo models.SessionRepository, activityRepo models.ActivityRepository, metricsRepo models.MetricsRepository) *MonitoringService {
	ms := &MonitoringService{
		sessionRepo:    sessionRepo,
		activityRepo:   activityRepo,
		metricsRepo:    metricsRepo,
		requestMetrics: &RequestMetrics{},
	}
	
	// Start background cleanup routine
	go ms.runBackgroundTasks()
	
	return ms
}

// CreateUserSession creates a new user session and logs the login activity
func (ms *MonitoringService) CreateUserSession(userID int, username, ipAddress, userAgent string) (*models.UserSession, error) {
	log.Printf("MonitoringService: Creating session for user %d (%s) from IP %s", userID, username, ipAddress)
	
	session, err := ms.sessionRepo.CreateSession(userID, ipAddress, userAgent)
	if err != nil {
		log.Printf("MonitoringService: Failed to create session: %v", err)
		ms.LogActivity(userID, username, "login", "Session creation failed", ipAddress, false, err.Error())
		return nil, err
	}
	
	log.Printf("MonitoringService: Session %d created successfully for user %d", session.ID, userID)
	ms.LogActivity(userID, username, "login", fmt.Sprintf("Session %d created", session.ID), ipAddress, true, "")
	return session, nil
}

// FindOrCreateUserSession finds an existing active session or creates a new one
func (ms *MonitoringService) FindOrCreateUserSession(userID int, username, ipAddress, userAgent string) (*models.UserSession, error) {
	log.Printf("MonitoringService: Finding or creating session for user %d (%s) from IP %s", userID, username, ipAddress)
	
	session, err := ms.sessionRepo.FindOrCreateActiveSession(userID, ipAddress, userAgent)
	if err != nil {
		log.Printf("MonitoringService: Failed to find/create session: %v", err)
		ms.LogActivity(userID, username, "session_error", "Session find/create failed", ipAddress, false, err.Error())
		return nil, err
	}
	
	// If this was a new session, log the login activity
	if session.LoginTime.After(time.Now().Add(-5*time.Second)) {
		log.Printf("MonitoringService: New session %d created for user %d", session.ID, userID)
		ms.LogActivity(userID, username, "login", fmt.Sprintf("Session %d created", session.ID), ipAddress, true, "")
	} else {
		log.Printf("MonitoringService: Existing session %d updated for user %d", session.ID, userID)
	}
	
	return session, nil
}

// EndUserSession ends a user session and logs the logout activity
func (ms *MonitoringService) EndUserSession(sessionID int, userID int, username, ipAddress string) error {
	err := ms.sessionRepo.EndSession(sessionID)
	if err != nil {
		ms.LogActivity(userID, username, "logout", "Session end failed", ipAddress, false, err.Error())
		return err
	}
	
	ms.LogActivity(userID, username, "logout", fmt.Sprintf("Session %d ended", sessionID), ipAddress, true, "")
	return nil
}

// LogActivity records user activity with error handling
func (ms *MonitoringService) LogActivity(userID int, username, action, details, ipAddress string, success bool, errorMsg string) {
	err := ms.activityRepo.LogActivity(userID, username, action, details, ipAddress, success, errorMsg)
	if err != nil {
		log.Printf("Failed to log activity: %v", err)
	}
}

// LogTrade records a trade activity and updates session trade count
func (ms *MonitoringService) LogTrade(sessionID, userID int, username, symbol, action string, quantity int, price float64, ipAddress string, success bool, errorMsg string) {
	details := fmt.Sprintf("%s %d shares of %s at $%.2f", action, quantity, symbol, price)
	ms.LogActivity(userID, username, "trade", details, ipAddress, success, errorMsg)
	
	if success && sessionID > 0 {
		ms.sessionRepo.IncrementTradeCount(sessionID)
	}
}

// TrackWebSocketConnection adds a WebSocket connection to monitoring
func (ms *MonitoringService) TrackWebSocketConnection(connectionID string, userID int, username, ipAddress string) {
	ms.activeConnections.Store(connectionID, &ConnectionInfo{
		UserID:      userID,
		Username:    username,
		ConnectedAt: time.Now(),
		IPAddress:   ipAddress,
	})
	
	ms.LogActivity(userID, username, "websocket_connect", fmt.Sprintf("Connection %s established", connectionID), ipAddress, true, "")
}

// RemoveWebSocketConnection removes a WebSocket connection from monitoring
func (ms *MonitoringService) RemoveWebSocketConnection(connectionID string) {
	if connInfo, ok := ms.activeConnections.LoadAndDelete(connectionID); ok {
		if info, ok := connInfo.(*ConnectionInfo); ok {
			ms.LogActivity(info.UserID, info.Username, "websocket_disconnect", 
				fmt.Sprintf("Connection %s closed", connectionID), info.IPAddress, true, "")
		}
	}
}

// GetActiveWebSocketCount returns the number of active WebSocket connections
func (ms *MonitoringService) GetActiveWebSocketCount() int {
	count := 0
	ms.activeConnections.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// RecordRequest tracks API request metrics
func (ms *MonitoringService) RecordRequest(duration time.Duration, success bool) {
	ms.requestMetrics.mu.Lock()
	defer ms.requestMetrics.mu.Unlock()
	
	ms.requestMetrics.TotalRequests++
	ms.requestMetrics.TotalTime += duration
	
	now := time.Now()
	ms.requestMetrics.LastHourReqs = append(ms.requestMetrics.LastHourReqs, now)
	
	if !success {
		ms.requestMetrics.FailedRequests++
		ms.requestMetrics.LastHourErrors = append(ms.requestMetrics.LastHourErrors, now)
	}
	
	// Clean old entries (keep only last hour)
	hourAgo := now.Add(-time.Hour)
	ms.requestMetrics.LastHourReqs = ms.filterTimesAfter(ms.requestMetrics.LastHourReqs, hourAgo)
	ms.requestMetrics.LastHourErrors = ms.filterTimesAfter(ms.requestMetrics.LastHourErrors, hourAgo)
}

// GetSystemMetrics returns comprehensive system metrics
func (ms *MonitoringService) GetSystemMetrics() (*models.SystemMetrics, error) {
	metrics, err := ms.metricsRepo.GetSystemMetrics()
	if err != nil {
		return nil, err
	}
	
	// Add real-time WebSocket connection count
	metrics.WebSocketConns = ms.GetActiveWebSocketCount()
	
	// Add real-time request metrics
	ms.requestMetrics.mu.RLock()
	if len(ms.requestMetrics.LastHourReqs) > 0 {
		avgTime := ms.requestMetrics.TotalTime.Milliseconds() / ms.requestMetrics.TotalRequests
		metrics.AvgResponseTime = float64(avgTime)
		
		errorRate := float64(len(ms.requestMetrics.LastHourErrors)) / float64(len(ms.requestMetrics.LastHourReqs)) * 100
		metrics.ErrorRate = errorRate
	}
	ms.requestMetrics.mu.RUnlock()
	
	return metrics, nil
}

// GetActiveSessions returns all currently active user sessions
func (ms *MonitoringService) GetActiveSessions() ([]*models.UserSession, error) {
	return ms.sessionRepo.GetActiveSessions()
}

// GetRecentActivity returns recent user activity
func (ms *MonitoringService) GetRecentActivity(limit int) ([]*models.UserActivity, error) {
	return ms.activityRepo.GetRecentActivity(limit)
}

// GetUserActivity returns activity for a specific user
func (ms *MonitoringService) GetUserActivity(userID int, limit int) ([]*models.UserActivity, error) {
	return ms.activityRepo.GetUserActivity(userID, limit)
}

// GetUserSessions returns recent sessions for a specific user
func (ms *MonitoringService) GetUserSessions(userID int, limit int) ([]*models.UserSession, error) {
	return ms.sessionRepo.GetUserSessions(userID, limit)
}

// GetActivityByTimeRange returns activity within a time range
func (ms *MonitoringService) GetActivityByTimeRange(startTime, endTime time.Time) ([]*models.UserActivity, error) {
	return ms.activityRepo.GetActivityByTimeRange(startTime, endTime)
}

// CreateRequestTrackerMiddleware returns HTTP middleware for request tracking
func (ms *MonitoringService) CreateRequestTrackerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a response writer wrapper to capture status
			ww := &responseWriterWrapper{ResponseWriter: w, statusCode: 200}
			
			next.ServeHTTP(ww, r)
			
			duration := time.Since(start)
			success := ww.statusCode < 400
			
			ms.RecordRequest(duration, success)
		})
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture status codes
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Helper methods

func (ms *MonitoringService) filterTimesAfter(times []time.Time, after time.Time) []time.Time {
	var filtered []time.Time
	for _, t := range times {
		if t.After(after) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// Background tasks
func (ms *MonitoringService) runBackgroundTasks() {
	// Cleanup expired sessions every 10 minutes
	cleanupTicker := time.NewTicker(10 * time.Minute)
	defer cleanupTicker.Stop()
	
	// Record metrics every 5 minutes
	metricsTicker := time.NewTicker(5 * time.Minute)
	defer metricsTicker.Stop()
	
	for {
		select {
		case <-cleanupTicker.C:
			ms.cleanupExpiredSessions()
		case <-metricsTicker.C:
			ms.recordSystemMetrics()
		}
	}
}

func (ms *MonitoringService) cleanupExpiredSessions() {
	err := ms.sessionRepo.CleanupExpiredSessions()
	if err != nil {
		log.Printf("Failed to cleanup expired sessions: %v", err)
	}
}

func (ms *MonitoringService) recordSystemMetrics() {
	if metricsRepo, ok := ms.metricsRepo.(*repository.MetricsRepo); ok {
		err := metricsRepo.RecordSystemMetrics()
		if err != nil {
			log.Printf("Failed to record system metrics: %v", err)
		}
	}
}

// Shutdown gracefully shuts down the monitoring service
func (ms *MonitoringService) Shutdown(ctx context.Context) error {
	// End all active sessions
	sessions, err := ms.GetActiveSessions()
	if err == nil {
		for _, session := range sessions {
			ms.sessionRepo.EndSession(session.ID)
		}
	}
	
	return nil
}