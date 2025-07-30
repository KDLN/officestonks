package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"officestonks/internal/models"
	"officestonks/internal/services"
)

// ChangelogHandler handles changelog-related HTTP requests
type ChangelogHandler struct {
	changelogService *services.ChangelogService
}

// NewChangelogHandler creates a new changelog handler
func NewChangelogHandler(changelogService *services.ChangelogService) *ChangelogHandler {
	return &ChangelogHandler{
		changelogService: changelogService,
	}
}

// CreateChangelogRequest represents the request body for creating changelog entries
type CreateChangelogRequest struct {
	Version     string              `json:"version"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Changes     []string            `json:"changes"`
	ChangeType  models.ChangeType   `json:"change_type"`
	IsMajor     bool                `json:"is_major"`
}

// GetPublicChangelog returns visible changelog entries for public display
func (h *ChangelogHandler) GetPublicChangelog(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	majorOnly := r.URL.Query().Get("major_only") == "true"

	limit := 20 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var entries []*models.ChangelogEntry
	var err error

	if majorOnly {
		entries, err = h.changelogService.GetMajorEntries(limit)
	} else {
		entries, err = h.changelogService.GetVisibleEntries(limit, offset)
	}

	if err != nil {
		http.Error(w, "Failed to fetch changelog entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}

// GetChangelogByVersion returns a specific changelog entry by version
func (h *ChangelogHandler) GetChangelogByVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars["version"]

	if version == "" {
		http.Error(w, "Version parameter is required", http.StatusBadRequest)
		return
	}

	entry, err := h.changelogService.GetEntryByVersion(version)
	if err != nil {
		http.Error(w, "Changelog entry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// CreateChangelog creates a new changelog entry (admin only)
func (h *ChangelogHandler) CreateChangelog(w http.ResponseWriter, r *http.Request) {
	var req CreateChangelogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Version == "" || req.Title == "" || len(req.Changes) == 0 {
		http.Error(w, "Version, title, and changes are required", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(int)
	var createdBy *int
	if ok {
		createdBy = &userID
	}

	entry, err := h.changelogService.CreateEntry(
		req.Version,
		req.Title,
		req.Description,
		req.Changes,
		req.ChangeType,
		req.IsMajor,
		createdBy,
	)

	if err != nil {
		http.Error(w, "Failed to create changelog entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// GetAllChangelog returns all changelog entries (admin only)
func (h *ChangelogHandler) GetAllChangelog(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	entries, err := h.changelogService.GetAllEntries(limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch changelog entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}

// UpdateChangelogVisibility updates the visibility of a changelog entry (admin only)
func (h *ChangelogHandler) UpdateChangelogVisibility(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	var req struct {
		IsVisible bool `json:"is_visible"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.changelogService.UpdateEntryVisibility(id, req.IsVisible)
	if err != nil {
		http.Error(w, "Failed to update changelog visibility", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// DeleteChangelog deletes a changelog entry (admin only)
func (h *ChangelogHandler) DeleteChangelog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	err = h.changelogService.DeleteEntry(id)
	if err != nil {
		http.Error(w, "Failed to delete changelog entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}