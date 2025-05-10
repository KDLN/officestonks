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
		// CRITICAL: Set CORS headers immediately, at the very top
		origin := r.Header.Get("Origin")

		// Always allow the production frontend origin unconditionally
		if origin == "https://officestonks-frontend-production.up.railway.app" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set all other CORS headers
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

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
			r.Header.Set("Authorization", "Bearer "+token)
			log.Printf("AdminOnly: Added token from URL parameter: %s", tokenPrefix)
		}

		// Get user ID from context (set by auth middleware)
		userID, ok := r.Context().Value("userID").(int)
		log.Printf("AdminOnly: UserID from context: %v, ok: %v", userID, ok)

		if !ok {
			log.Printf("AdminOnly: No userID in context, responding with 401")
			// Add CORS headers to error response
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user is admin
		isAdmin, err := h.userRepo.IsUserAdmin(userID)
		log.Printf("AdminOnly: User %d isAdmin=%v, err=%v", userID, isAdmin, err)

		if err != nil {
			log.Printf("AdminOnly: Error checking admin status: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !isAdmin {
			log.Printf("AdminOnly: User %d is not an admin, responding with 403", userID)
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		// User is admin, proceed
		log.Printf("AdminOnly: User %d authorized as admin, proceeding", userID)
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
	// Set CORS headers first, before anything else
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Use wildcard for debugging
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

	// Log request details for debugging
	log.Printf("GetAllUsers called with method: %s from origin: %s", r.Method, origin)
	log.Printf("GetAllUsers: Request headers: %v", r.Header)

	// Debug User ID and Admin Status
	userID, ok := r.Context().Value("userID").(int)
	if ok {
		log.Printf("GetAllUsers: User ID from context: %d", userID)
		isAdmin, err := h.userRepo.IsUserAdmin(userID)
		if err != nil {
			log.Printf("GetAllUsers: Error checking admin status: %v", err)
		} else {
			log.Printf("GetAllUsers: User is admin: %v", isAdmin)
		}
	} else {
		log.Printf("GetAllUsers: No user ID in context")
	}

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		log.Printf("GetAllUsers: Handling OPTIONS preflight request")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get all users
	log.Printf("GetAllUsers: Fetching users from repository")
	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		log.Printf("GetAllUsers: Error getting all users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("GetAllUsers: Repository returned %d users", len(users))

	w.Header().Set("Content-Type", "application/json")

	// Make sure valid JSON is sent
	if users == nil || len(users) == 0 {
		log.Printf("GetAllUsers: No users found, returning empty array")
		// Return empty array instead of null
		w.Write([]byte("[]"))
		return
	}

	// Debug user data
	for i, user := range users {
		log.Printf("GetAllUsers: User[%d]: id=%d, username=%s, isAdmin=%v",
			i, user.ID, user.Username, user.IsAdmin)
	}

	// Encode the response
	log.Printf("GetAllUsers: Encoding %d users to JSON", len(users))
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Printf("GetAllUsers: Error encoding users: %v", err)
		// Return empty array in case of encoding error
		w.Write([]byte("[]"))
	} else {
		log.Printf("GetAllUsers: Successfully encoded and sent users")
	}
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

// ResetStockPrices resets all stock prices to random values (admin only)
func (h *AdminHandler) ResetStockPrices(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers first, before anything else
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
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