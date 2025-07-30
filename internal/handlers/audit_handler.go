package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"officestonks/internal/services"
)

// AuditHandler exposes audit log endpoints
type AuditHandler struct {
	service *services.AuditService
}

func NewAuditHandler(service *services.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// GetRecentEvents returns recent audit log entries (admin only)
func (h *AuditHandler) GetRecentEvents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	events, err := h.service.GetRecentEvents(limit)
	if err != nil {
		http.Error(w, "Failed to fetch audit log", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
