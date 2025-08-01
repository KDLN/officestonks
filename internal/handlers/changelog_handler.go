package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"officestonks/internal/models"
	"officestonks/internal/services"
)

// ChangelogFile represents the structure of the changelog.json file
type ChangelogFile struct {
	Entries []*models.ChangelogEntry `json:"entries"`
}

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

// readChangelogFile reads the changelog from the JSON file
func (h *ChangelogHandler) readChangelogFile() ([]*models.ChangelogEntry, error) {
	// Try to find changelog.json in current directory or project root
	filePaths := []string{
		"changelog.json",
		"./changelog.json",
		"../changelog.json",
		"../../changelog.json",
	}
	
	var data []byte
	var err error
	
	for _, path := range filePaths {
		if _, statErr := os.Stat(path); statErr == nil {
			data, err = ioutil.ReadFile(path)
			if err == nil {
				log.Printf("📖 Reading changelog from: %s", path)
				break
			}
		}
	}
	
	if err != nil || len(data) == 0 {
		log.Printf("⚠️ Could not read changelog file, returning empty entries")
		return []*models.ChangelogEntry{}, nil
	}
	
	var changelogFile ChangelogFile
	if err := json.Unmarshal(data, &changelogFile); err != nil {
		log.Printf("❌ Error parsing changelog JSON: %v", err)
		return []*models.ChangelogEntry{}, nil
	}
	
	// Parse created_at timestamps if they're strings
	for _, entry := range changelogFile.Entries {
		if entry.CreatedAt.IsZero() {
			// If timestamp parsing failed, use current time
			entry.CreatedAt = time.Now()
		}
	}
	
	// Sort entries by ID descending (newest first)
	sort.Slice(changelogFile.Entries, func(i, j int) bool {
		return changelogFile.Entries[i].ID > changelogFile.Entries[j].ID
	})
	
	log.Printf("✅ Loaded %d changelog entries from file", len(changelogFile.Entries))
	return changelogFile.Entries, nil
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

	// Use database service instead of file
	allEntries, err := h.changelogService.GetVisibleEntries(limit*2, 0) // Get more entries to handle filtering
	if err != nil {
		log.Printf("Error fetching changelog from database: %v", err)
		http.Error(w, "Failed to fetch changelog entries", http.StatusInternalServerError)
		return
	}

	// Filter by major releases if requested
	var visibleEntries []*models.ChangelogEntry
	for _, entry := range allEntries {
		if !majorOnly || entry.IsMajor {
			visibleEntries = append(visibleEntries, entry)
		}
	}

	// Apply pagination
	start := offset
	if start > len(visibleEntries) {
		start = len(visibleEntries)
	}
	
	end := start + limit
	if end > len(visibleEntries) {
		end = len(visibleEntries)
	}
	
	entries := visibleEntries[start:end]

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

	// Read from file instead of database
	allEntries, err := h.readChangelogFile()
	if err != nil {
		http.Error(w, "Failed to fetch changelog entries", http.StatusInternalServerError)
		return
	}

	// Find entry by version
	var foundEntry *models.ChangelogEntry
	for _, entry := range allEntries {
		if entry.Version == version && entry.IsVisible {
			foundEntry = entry
			break
		}
	}

	if foundEntry == nil {
		http.Error(w, "Changelog entry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(foundEntry)
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