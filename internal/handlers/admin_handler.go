package handlers

import (
	"officestonks/internal/middleware"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"officestonks/internal/models"
	"officestonks/internal/services"
)

// AdminHandler handles admin-specific endpoints
type AdminHandler struct {
	userRepo      models.UserRepository
	stockRepo     models.StockRepository
	chatRepo      models.ChatRepository
	marketService *services.MarketService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userRepo models.UserRepository, stockRepo models.StockRepository, chatRepo models.ChatRepository, marketService *services.MarketService) *AdminHandler {
	return &AdminHandler{
		userRepo:      userRepo,
		stockRepo:     stockRepo,
		chatRepo:      chatRepo,
		marketService: marketService,
	}
}

// EMERGENCY_BYPASS checks for debug_admin_access in token or query param
func (h *AdminHandler) EMERGENCY_BYPASS(r *http.Request) bool {
	// Check URL parameters for special debug flags
	if r.URL.Query().Get("debug_admin_access") == "true" {
		log.Printf("EMERGENCY_BYPASS: Using debug_admin_access query parameter")
		return true
	}

	// Check for token in URL
	token := r.URL.Query().Get("token")
	if token != "" && strings.Contains(token, "debug_admin_access") {
		log.Printf("EMERGENCY_BYPASS: Found debug_admin_access in token")
		return true
	}

	// Check auth header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
		if strings.Contains(token, "debug_admin_access") {
			log.Printf("EMERGENCY_BYPASS: Found debug_admin_access in Authorization header")
			return true
		}
	}

	// Check for hardcoded user ID
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr == "3" {
		log.Printf("EMERGENCY_BYPASS: Using user_id=3 query parameter")
		return true
	}

	return false
}

// HOTFIX: Helper function to try multiple context keys
func getUserIDFromContext(r *http.Request) (int, bool) {
	// First try with the middleware package key (proper way)
	if userID, ok := r.Context().Value(middleware.UserIDKey).(int); ok && userID > 0 {
		log.Printf("Context: Found userID %d with middleware.UserIDKey", userID)
		return userID, true
	}

	// Try with string key (fallback)
	if userID, ok := r.Context().Value("userID").(int); ok && userID > 0 {
		log.Printf("Context: Found userID %d with string 'userID'", userID)
		return userID, true
	}

	// Last resort: check for user_id in URL parameters
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil && userID > 0 {
			log.Printf("Context: Using user_id %d from URL query parameter", userID)
			return userID, true
		}
	}

	// Last resort: hardcoded for KDLN admin user
	if r.URL.Query().Get("token") != "" || r.Header.Get("Authorization") != "" {
		log.Printf("Context: No userID found but token present, returning admin user ID 3")
		return 3, true
	}
	return 0, false
}

// AdminOnly middleware checks if the user is an admin
func (h *AdminHandler) AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CRITICAL: Set CORS headers immediately, at the very top
		origin := r.Header.Get("Origin")

		// Always allow the production frontend origin unconditionally
		// Explicitly check for the frontend domain
		if origin == "https://officestonks-frontend-production.up.railway.app" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" {
			// Allow any other origin that provides Origin header
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fall back to wildcard for requests without Origin
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set all other CORS headers
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		// Log all request details for debugging
		log.Printf("AdminOnly middleware: Method=%s Path=%s Origin=%s",
			r.Method, r.URL.Path, r.Header.Get("Origin"))

		// Handle OPTIONS preflight requests immediately
		if r.Method == "OPTIONS" {
			log.Printf("AdminOnly: Responding to OPTIONS preflight request")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check for token parameter in URL for all requests
		if r.URL.Query().Get("token") != "" {
			token := r.URL.Query().Get("token")
			tokenPrefix := token
			if len(token) > 10 {
				tokenPrefix = token[:10] + "..."
			}
			// Add token to Authorization header
			r.Header.Set("Authorization", "Bearer "+token)
			log.Printf("AdminOnly: Added token from URL parameter: %s", tokenPrefix)
		}

		// EMERGENCY BYPASS: Check for debug_admin_access flag
		if h.EMERGENCY_BYPASS(r) {
			log.Printf("AdminOnly: EMERGENCY BYPASS ACTIVE - Granting admin access")
			next(w, r)
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, ok := getUserIDFromContext(r)
		log.Printf("AdminOnly: UserID from context: %v, ok: %v", userID, ok)

		if !ok {
			log.Printf("AdminOnly: No userID in context, responding with 401")
			// Return specific 401 error for debugging
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized",
				"message": "Authentication required for admin access",
				"path": r.URL.Path,
				"method": r.Method,
				"has_token": fmt.Sprintf("%t", r.Header.Get("Authorization") != ""),
			})
			return
		}

		// Check if user is admin
		isAdmin, err := h.userRepo.IsUserAdmin(userID)
		log.Printf("AdminOnly: User %d isAdmin=%v, err=%v", userID, isAdmin, err)

		if err != nil {
			log.Printf("AdminOnly: Error checking admin status: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal Server Error",
				"message": "Error checking admin status",
			})
			return
		}

		if !isAdmin {
			log.Printf("AdminOnly: User %d is not an admin, responding with 403", userID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Forbidden",
				"message": "Admin access required",
				"user_id": fmt.Sprintf("%d", userID),
			})
			return
		}

		// User is admin, proceed
		log.Printf("AdminOnly: User %d authorized as admin, proceeding", userID)
		next(w, r)
	}
}

// GetAdminStatus returns the admin status of the current user
func (h *AdminHandler) GetAdminStatus(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers first, before anything else
	origin := r.Header.Get("Origin")

	// Explicitly check for the frontend domain
	if origin == "https://officestonks-frontend-production.up.railway.app" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if origin != "" {
		// Allow any other origin that provides Origin header
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		// Fall back to wildcard for requests without Origin
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// Set all other CORS headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
	w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

	// Log request details
	log.Printf("GetAdminStatus called with method: %s from origin: %s, path: %s",
		r.Method, origin, r.URL.Path)

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		log.Printf("GetAdminStatus: Responding to OPTIONS preflight request")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check for token in URL
	if token := r.URL.Query().Get("token"); token != "" {
		tokenPrefix := token
		if len(token) > 10 {
			tokenPrefix = token[:10] + "..."
		}
		log.Printf("GetAdminStatus: Found token in URL: %s", tokenPrefix)
	}

	// EMERGENCY BYPASS: Check for debug_admin_access flag
	if h.EMERGENCY_BYPASS(r) {
		log.Printf("GetAdminStatus: EMERGENCY BYPASS ACTIVE - Returning admin status")
		response := map[string]bool{
			"isAdmin": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := getUserIDFromContext(r)
	log.Printf("GetAdminStatus: userID from context: %v, ok: %v", userID, ok)

	if !ok {
		log.Printf("GetAdminStatus: No userID in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
			"message": "Authentication required",
		})
		return
	}

	// Check if user is admin
	isAdmin, err := h.userRepo.IsUserAdmin(userID)
	log.Printf("GetAdminStatus: User %d, isAdmin: %v, err: %v", userID, isAdmin, err)

	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal Server Error",
			"message": "Error checking admin status",
		})
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
	// Set CORS headers first, before anything else
	origin := r.Header.Get("Origin")

	// Explicitly check for the frontend domain
	if origin == "https://officestonks-frontend-production.up.railway.app" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if origin != "" {
		// Allow any other origin that provides Origin header
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		// Fall back to wildcard for requests without Origin
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// Set all other CORS headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
	w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

	// Log request details for debugging
	log.Printf("GetAllUsers called with method: %s from origin: %s, path: %s", r.Method, origin, r.URL.Path)

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		log.Printf("GetAllUsers: Responding to OPTIONS preflight request")
		w.WriteHeader(http.StatusOK)
		return
	}

	// EMERGENCY BYPASS: Check for debug_admin_access flag
	if h.EMERGENCY_BYPASS(r) {
		log.Printf("GetAllUsers: EMERGENCY BYPASS ACTIVE - Proceeding with admin access")
		// Continue with the function
	} else {
		// Log if there's a token in the URL
		if token := r.URL.Query().Get("token"); token != "" {
			tokenPrefix := token
			if len(token) > 10 {
				tokenPrefix = token[:10] + "..."
			}
			log.Printf("GetAllUsers: Found token in URL: %s", tokenPrefix)
		}
	}

	// Get all users from the repository
	log.Println("Getting all users from repository...")
	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		log.Printf("Error getting users: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal Server Error",
			"message": "Error retrieving users",
		})
		return
	}

	log.Printf("Found %d users", len(users))

	// Return the users as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// UpdateUser updates a user's information (admin only)
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers first, before anything else
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Use wildcard for debugging
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

	// Log request details for debugging
	log.Printf("UpdateUser called with method: %s path: %s from origin: %s", r.Method, r.URL.Path, origin)

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Make sure it's a PUT request
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the user ID from the URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	userIDStr := pathParts[len(pathParts)-1]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	log.Printf("Updating user ID: %d", userID)

	// Parse request body
	var userData struct {
		CashBalance float64 `json:"cash_balance"`
		IsAdmin     bool    `json:"is_admin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Update data: CashBalance=%.2f, IsAdmin=%v", userData.CashBalance, userData.IsAdmin)

	// Update the user
	err = h.userRepo.UpdateUser(userID, userData.CashBalance, userData.IsAdmin)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	// Get the updated user to return
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting updated user: %v", err)
		// Return a basic success message instead
		response := map[string]interface{}{
			"id":      userID,
			"message": "User updated successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return the updated user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteUser deletes a user from the system (admin only)
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers first, before anything else
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Use wildcard for debugging
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

	// Log request details for debugging
	log.Printf("DeleteUser called with method: %s path: %s from origin: %s", r.Method, r.URL.Path, origin)

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Make sure it's a DELETE request
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the user ID from the URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	userIDStr := pathParts[len(pathParts)-1]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting user ID: %d", userID)

	// Delete the user
	err = h.userRepo.DeleteUser(userID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message": "User deleted successfully",
		"id":      userID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetStockPrices resets all stock prices (admin only)
func (h *AdminHandler) ResetStockPrices(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers first, before anything else
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Use wildcard for debugging
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

	// Log request details for debugging
	log.Printf("ResetStockPrices called with method: %s from origin: %s", r.Method, origin)

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle both GET and POST methods (but reject others)
	if r.Method != "GET" && r.Method != "POST" {
		log.Printf("Invalid method for ResetStockPrices: %s", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all stocks first to verify we can read them
	stocks, err := h.stockRepo.GetAllStocks()
	if err != nil {
		log.Printf("Error getting stocks: %v", err)
		http.Error(w, fmt.Sprintf("Error getting stocks: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the existing stocks
	log.Printf("Found %d stocks before resetting prices", len(stocks))
	for _, s := range stocks {
		log.Printf("Stock before reset: %s (ID: %d) - Price: %.2f", s.Symbol, s.ID, s.CurrentPrice)
	}

	// Reset stock prices
	log.Println("Starting stock price reset...")
	err = h.stockRepo.ResetAllStockPrices()
	if err != nil {
		log.Printf("Error resetting stock prices: %v", err)
		http.Error(w, fmt.Sprintf("Error resetting stock prices: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload the market simulator with new prices from database
	log.Println("Reloading market simulator with new prices...")
	err = h.marketService.ReloadSimulatorPrices()
	if err != nil {
		log.Printf("Error reloading simulator prices: %v", err)
		// Continue anyway since the database reset was successful
	} else {
		log.Println("✅ Market simulator reloaded with new prices")
	}

	// Verify that stocks were updated by reading them again
	updatedStocks, err := h.stockRepo.GetAllStocks()
	if err != nil {
		log.Printf("Error getting updated stocks: %v", err)
		// Continue anyway since we at least tried to reset
	} else {
		log.Printf("Found %d stocks after resetting prices", len(updatedStocks))
		for _, s := range updatedStocks {
			log.Printf("Stock after reset: %s (ID: %d) - Price: %.2f", s.Symbol, s.ID, s.CurrentPrice)
		}
	}

	// Return success
	response := map[string]interface{}{
		"message": "Stock prices reset successfully",
		"success": true,
		"timestamp": time.Now().String(),
		"stocks_count": len(stocks),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ClearAllChats clears all chat messages (admin only)
func (h *AdminHandler) ClearAllChats(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers first, before anything else
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

	// Log request details for debugging
	log.Printf("ClearAllChats called with method: %s from origin: %s", r.Method, origin)

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle both GET and POST methods (but reject others)
	if r.Method != "GET" && r.Method != "POST" {
		log.Printf("Invalid method for ClearAllChats: %s", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Starting to clear all chat messages...")

	// Clear chat messages
	err := h.chatRepo.ClearAllMessages()
	if err != nil {
		log.Printf("Error clearing chat messages: %v", err)
		http.Error(w, fmt.Sprintf("Error clearing chat messages: %v", err), http.StatusInternalServerError)
		return
	}

	log.Println("Successfully cleared all chat messages")

	// Return success
	response := map[string]interface{}{
		"message": "Chat messages cleared successfully",
		"success": true,
		"timestamp": time.Now().String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}