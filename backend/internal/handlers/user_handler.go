package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"officestonks/internal/middleware"
	"officestonks/internal/models"
	"officestonks/internal/services"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetLeaderboard returns the top users by portfolio value
func (h *UserHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Get the limit parameter, default to 10
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default value
	
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	// Get the leaderboard data
	leaderboard, err := h.userService.GetLeaderboard(limit)
	if err != nil {
		http.Error(w, "Failed to retrieve leaderboard", http.StatusInternalServerError)
		return
	}
	
	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

// GetUserProfile returns the profile for the authenticated user
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the middleware
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get the user profile
	userProfile, err := h.userService.GetUserProfile(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user profile", http.StatusInternalServerError)
		return
	}
	
	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userProfile)
}