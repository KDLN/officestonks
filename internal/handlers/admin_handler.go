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


// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(r *http.Request) (int, bool) {
	// Try with the middleware package key (proper way)
	if userID, ok := r.Context().Value(middleware.UserIDKey).(int); ok && userID > 0 {
		log.Printf("Context: Found userID %d with middleware.UserIDKey", userID)
		return userID, true
	}

	// Try with string key (fallback for compatibility)
	if userID, ok := r.Context().Value("userID").(int); ok && userID > 0 {
		log.Printf("Context: Found userID %d with string 'userID'", userID)
		return userID, true
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

	// Log if there's a token in the URL
	if token := r.URL.Query().Get("token"); token != "" {
		tokenPrefix := token
		if len(token) > 10 {
			tokenPrefix = token[:10] + "..."
		}
		log.Printf("GetAllUsers: Found token in URL: %s", tokenPrefix)
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

	// Perform atomic reset of stock prices (pause simulator, reset DB, reload simulator, resume)
	log.Println("Starting atomic stock price reset...")
	err = h.marketService.AtomicResetStockPrices()
	if err != nil {
		log.Printf("Error during atomic stock price reset: %v", err)
		http.Error(w, fmt.Sprintf("Error resetting stock prices: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("✅ Atomic stock price reset completed successfully")

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
	// Validate and fix any corrupted data in the market simulator
	log.Println("Validating market simulator after price reset...")
	h.marketService.ValidateSimulator()

	response := map[string]interface{}{
		"message": "Stock prices reset and market simulator reinitialized successfully",
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

// Crisis testing endpoints

// ForceCrisisEvent forces a stock into crisis for testing (admin only)
func (h *AdminHandler) ForceCrisisEvent(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse request body
	var req struct {
		StockID int `json:"stock_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.StockID <= 0 {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}
	
	log.Printf("🧪 Admin forcing crisis event for stock ID: %d", req.StockID)
	
	// Force crisis event
	err := h.marketService.ForceCrisisEvent(req.StockID)
	if err != nil {
		log.Printf("Error forcing crisis event: %v", err)
		http.Error(w, fmt.Sprintf("Error forcing crisis: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Get stock info after crisis
	stockInfo, err := h.marketService.GetSimulatorStockInfo(req.StockID)
	if err != nil {
		log.Printf("Error getting stock info: %v", err)
		stockInfo = map[string]interface{}{"error": "Could not retrieve updated stock info"}
	}
	
	response := map[string]interface{}{
		"message":    "Crisis event forced successfully",
		"success":    true,
		"stock_id":   req.StockID,
		"stock_info": stockInfo,
		"timestamp":  time.Now().String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ForceBankruptcy forces a stock into bankruptcy for testing (admin only)
func (h *AdminHandler) ForceBankruptcy(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		StockID int `json:"stock_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.StockID <= 0 {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}
	
	log.Printf("🧪 Admin forcing bankruptcy for stock ID: %d", req.StockID)
	
	err := h.marketService.ForceBankruptcy(req.StockID)
	if err != nil {
		log.Printf("Error forcing bankruptcy: %v", err)
		http.Error(w, fmt.Sprintf("Error forcing bankruptcy: %v", err), http.StatusInternalServerError)
		return
	}
	
	stockInfo, err := h.marketService.GetSimulatorStockInfo(req.StockID)
	if err != nil {
		log.Printf("Error getting stock info: %v", err)
		stockInfo = map[string]interface{}{"error": "Could not retrieve updated stock info"}
	}
	
	response := map[string]interface{}{
		"message":    "Bankruptcy forced successfully",
		"success":    true,
		"stock_id":   req.StockID,
		"stock_info": stockInfo,
		"timestamp":  time.Now().String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ForceRecovery forces a stock recovery for testing (admin only)
func (h *AdminHandler) ForceRecovery(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		StockID int `json:"stock_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.StockID <= 0 {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}
	
	log.Printf("🧪 Admin forcing recovery for stock ID: %d", req.StockID)
	
	err := h.marketService.ForceRecovery(req.StockID)
	if err != nil {
		log.Printf("Error forcing recovery: %v", err)
		http.Error(w, fmt.Sprintf("Error forcing recovery: %v", err), http.StatusInternalServerError)
		return
	}
	
	stockInfo, err := h.marketService.GetSimulatorStockInfo(req.StockID)
	if err != nil {
		log.Printf("Error getting stock info: %v", err)
		stockInfo = map[string]interface{}{"error": "Could not retrieve updated stock info"}
	}
	
	response := map[string]interface{}{
		"message":    "Recovery forced successfully",
		"success":    true,
		"stock_id":   req.StockID,
		"stock_info": stockInfo,
		"timestamp":  time.Now().String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSimulatorStatus returns the current status of all stocks in the simulator (admin only)
func (h *AdminHandler) GetSimulatorStatus(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	log.Println("🧪 Admin requesting simulator status")
	
	stocks := h.marketService.ListSimulatorStocks()
	
	response := map[string]interface{}{
		"message":     "Simulator status retrieved successfully",
		"success":     true,
		"stocks":      stocks,
		"stock_count": len(stocks),
		"timestamp":   time.Now().String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllStocksDetailed returns all stocks with full details for admin management
func (h *AdminHandler) GetAllStocksDetailed(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	stocks, err := h.stockRepo.GetAllStocks()
	if err != nil {
		log.Printf("Error getting stocks: %v", err)
		http.Error(w, "Failed to get stocks", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stocks)
}

// CreateStock creates a new stock
func (h *AdminHandler) CreateStock(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Symbol            string  `json:"symbol"`
		Name              string  `json:"name"`
		Sector            string  `json:"sector"`
		SectorID          int     `json:"sector_id"`
		InitialPrice      float64 `json:"initial_price"`
		MarketCapCategory string  `json:"market_cap_category"`
		VolatilityProfile string  `json:"volatility_profile"`
		Description       string  `json:"description"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.Symbol == "" || req.Name == "" || req.InitialPrice <= 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	
	// Create the stock
	stock, err := h.stockRepo.CreateStock(
		req.Symbol, req.Name, req.Sector, req.SectorID,
		req.InitialPrice, req.MarketCapCategory, req.VolatilityProfile, req.Description,
	)
	if err != nil {
		log.Printf("Error creating stock: %v", err)
		http.Error(w, "Failed to create stock", http.StatusInternalServerError)
		return
	}
	
	// Add to market simulator
	h.marketService.AddStockToSimulator(stock)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stock)
}

// UpdateStockAdmin updates stock details
func (h *AdminHandler) UpdateStockAdmin(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract stock ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	
	stockIDStr := pathParts[len(pathParts)-1]
	stockID, err := strconv.Atoi(stockIDStr)
	if err != nil {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Name              string  `json:"name"`
		Sector            string  `json:"sector"`
		SectorID          int     `json:"sector_id"`
		CurrentPrice      float64 `json:"current_price"`
		VolatilityProfile string  `json:"volatility_profile"`
		Description       string  `json:"description"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	log.Printf("🔬 Admin stock update - ID: %d, Request: %+v", stockID, req)
	
	// Get existing stock data to preserve values not being updated
	existingStock, err := h.stockRepo.GetStockByID(stockID)
	if err != nil {
		log.Printf("❌ Failed to get existing stock: %v", err)
		http.Error(w, "Stock not found", http.StatusNotFound)
		return
	}
	
	// Use existing values for fields that are empty/default in the request
	name := req.Name
	if name == "" {
		name = existingStock.Name
	}
	
	sector := req.Sector
	if sector == "" {
		sector = existingStock.Sector
	}
	
	sectorID := req.SectorID
	if sectorID <= 0 {
		// Use existing sector_id, or default to 1 if existing is also invalid
		if existingStock.SectorID != nil && *existingStock.SectorID > 0 {
			sectorID = *existingStock.SectorID
		} else {
			sectorID = 1 // Default to Technology
		}
		log.Printf("🔧 Using preserved/default sector_id: %d", sectorID)
	}
	
	volatilityProfile := req.VolatilityProfile
	if volatilityProfile == "" {
		volatilityProfile = "normal" // Default value
	}
	
	description := req.Description
	// Description can be empty, so we allow it
	
	log.Printf("🔬 Final update parameters: name=%s, sector=%s, sectorID=%d, volatility=%s", 
		name, sector, sectorID, volatilityProfile)
	
	// Update stock details
	err = h.stockRepo.UpdateStockDetails(stockID, name, sector, sectorID, volatilityProfile, description)
	if err != nil {
		log.Printf("❌ UpdateStockDetails failed: %v", err)
		http.Error(w, "Failed to update stock", http.StatusInternalServerError)
		return
	}
	
	// Update price if provided
	if req.CurrentPrice > 0 {
		err = h.stockRepo.UpdateStockPrice(stockID, req.CurrentPrice)
		if err != nil {
			log.Printf("Error updating stock price: %v", err)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Stock updated successfully"})
}

// DeleteStockAdmin soft deletes a stock
func (h *AdminHandler) DeleteStockAdmin(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract stock ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	
	stockIDStr := pathParts[len(pathParts)-1]
	stockID, err := strconv.Atoi(stockIDStr)
	if err != nil {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}
	
	// Force delisting
	err = h.stockRepo.ForceDelisting(stockID, "Admin deletion")
	if err != nil {
		log.Printf("Error deleting stock: %v", err)
		http.Error(w, "Failed to delete stock", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Stock deleted successfully"})
}

// LaunchIPO launches a new stock as an IPO
func (h *AdminHandler) LaunchIPO(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Symbol          string  `json:"symbol"`
		Name            string  `json:"name"`
		Sector          string  `json:"sector"`
		SectorID        int     `json:"sector_id"`
		IPOPrice        float64 `json:"ipo_price"`
		SharesAvailable int     `json:"shares_available"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Default shares if not provided
	if req.SharesAvailable == 0 {
		req.SharesAvailable = 1000000
	}
	
	// Launch IPO
	stock, err := h.stockRepo.LaunchIPO(
		req.Symbol, req.Name, req.Sector, req.SectorID,
		req.IPOPrice, req.SharesAvailable,
	)
	if err != nil {
		log.Printf("Error launching IPO: %v", err)
		http.Error(w, "Failed to launch IPO", http.StatusInternalServerError)
		return
	}
	
	// Add to market simulator
	h.marketService.AddStockToSimulator(stock)
	
	// Create news event for IPO
	h.marketService.GenerateIPONews(stock)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stock)
}

// TriggerSectorEvent triggers a sector-wide market event
func (h *AdminHandler) TriggerSectorEvent(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		SectorID         int     `json:"sector_id"`
		EventType        string  `json:"event_type"` // "boom" or "crash"
		ImpactPercentage float64 `json:"impact_percentage"`
		DurationMinutes  int     `json:"duration_minutes"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Apply sector-wide impact
	err := h.marketService.ApplySectorEvent(req.SectorID, req.EventType, req.ImpactPercentage, req.DurationMinutes)
	if err != nil {
		log.Printf("Error applying sector event: %v", err)
		http.Error(w, "Failed to apply sector event", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Sector event triggered successfully"})
}

// setCORSHeaders helper function to set CORS headers consistently
func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
}