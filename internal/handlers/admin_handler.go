package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	
	"officestonks/internal/models"
)

// AdminHandler handles admin-specific endpoints
type AdminHandler struct {
	userRepo    models.UserRepository
	stockRepo   models.StockRepository
	chatRepo    models.ChatRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userRepo models.UserRepository, stockRepo models.StockRepository, chatRepo models.ChatRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:  userRepo,
		stockRepo: stockRepo,
		chatRepo:  chatRepo,
	}
}

// AdminOnly middleware checks if the user is an admin
func (h *AdminHandler) AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by auth middleware)
		userID, ok := r.Context().Value("userID").(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Check if user is admin
		isAdmin, err := h.userRepo.IsUserAdmin(userID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		if !isAdmin {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}
		
		// User is admin, proceed
		next(w, r)
	}
}

// GetAdminStatus returns the admin status of the current user
func (h *AdminHandler) GetAdminStatus(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(int)
	log.Printf("GetAdminStatus: userID from context: %v, ok: %v", userID, ok)

	if !ok {
		log.Printf("GetAdminStatus: No userID in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Debug admin status
	debugInfo := h.userRepo.DebugIsUserAdmin(userID)
	log.Printf("GetAdminStatus: Debug info: %s", debugInfo)

	// Check if user is admin
	isAdmin, err := h.userRepo.IsUserAdmin(userID)
	log.Printf("GetAdminStatus: User %d, isAdmin: %v, err: %v", userID, isAdmin, err)

	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return admin status
	response := map[string]bool{
		"isAdmin": isAdmin,
	}

	log.Printf("GetAdminStatus: Returning response for user %d: %v", userID, response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllUsers returns all users in the system (admin only)
func (h *AdminHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// UpdateUser updates a user's information (admin only)
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	userID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var updateRequest struct {
		Username    string  `json:"username"`
		CashBalance float64 `json:"cash_balance"`
		IsAdmin     bool    `json:"is_admin"`
	}
	
	err = json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Check if user exists
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Update user
	err = h.userRepo.UpdateUser(userID, updateRequest.CashBalance, updateRequest.IsAdmin)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Return updated user
	user.CashBalance = updateRequest.CashBalance
	user.IsAdmin = updateRequest.IsAdmin
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteUser deletes a user from the system (admin only)
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	userID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	// Delete user
	err = h.userRepo.DeleteUser(userID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Return success
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "User deleted successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// ResetStockPrices resets all stock prices to their initial values (admin only)
func (h *AdminHandler) ResetStockPrices(w http.ResponseWriter, r *http.Request) {
	// Log request method for debugging
	log.Printf("ResetStockPrices called with method: %s", r.Method)

	// Handle both GET and POST methods
	if r.Method != "GET" && r.Method != "POST" {
		log.Printf("Invalid method for ResetStockPrices: %s", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Reset stock prices
	err := h.stockRepo.ResetAllStockPrices()
	if err != nil {
		log.Printf("Error resetting stock prices: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success
	response := map[string]string{
		"message": "Stock prices reset successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	json.NewEncoder(w).Encode(response)
}

// ClearAllChats clears all chat messages (admin only)
func (h *AdminHandler) ClearAllChats(w http.ResponseWriter, r *http.Request) {
	// Clear chat messages
	err := h.chatRepo.ClearAllMessages()
	if err != nil {
		log.Printf("Error clearing chat messages: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Return success
	response := map[string]string{
		"message": "Chat messages cleared successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}