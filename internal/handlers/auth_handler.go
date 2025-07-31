package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"officestonks/internal/auth"
	"officestonks/internal/models"
	"officestonks/internal/services"
)

// getClientIP attempts to determine the real client IP address accounting for
// reverse proxies and load balancers.
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService  *services.AuthService
	auditService *services.AuditService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService, auditService *services.AuditService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		auditService: auditService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Register user
	resp, err := h.authService.Register(req.Username, req.Password)
	if err != nil {
		log.Printf("Error registering user: %v", err)
		// Return a more specific error message
		if err.Error() == "username already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Error registering user", http.StatusInternalServerError)
		}
		return
	}

	if h.auditService != nil {
		ip := getClientIP(r)
		_ = h.auditService.LogEvent(resp.UserID, "register", ip)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Login user
	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		log.Printf("Error logging in user: %v", err)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if h.auditService != nil {
		ip := getClientIP(r)
		_ = h.auditService.LogEvent(resp.UserID, "login", ip)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout handles user logout (no server-side state to clear currently)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// JWT-based auth doesn't require server-side logout
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"message": "Logged out successfully"}
	json.NewEncoder(w).Encode(resp)
}

// SupabaseAuth handles Supabase authentication
func (h *AuthHandler) SupabaseAuth(w http.ResponseWriter, r *http.Request) {
	// Get authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
		return
	}

	// Validate Supabase token
	supabaseClaims, err := auth.ValidateSupabaseToken(token)
	if err != nil {
		log.Printf("Supabase token validation error: %v", err)
		http.Error(w, "Invalid Supabase token", http.StatusUnauthorized)
		return
	}

	// Get or create user based on Supabase claims
	userID, err := h.authService.GetOrCreateSupabaseUser(supabaseClaims)
	if err != nil {
		log.Printf("Error getting/creating Supabase user: %v", err)
		http.Error(w, "Error processing Supabase user", http.StatusInternalServerError)
		return
	}

	// Generate Office Stonks JWT token
	officeToken, err := auth.GenerateToken(userID)
	if err != nil {
		log.Printf("Error generating Office Stonks token: %v", err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Get user details
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user details: %v", err)
		http.Error(w, "Error getting user details", http.StatusInternalServerError)
		return
	}

	// Return response
	resp := models.AuthResponse{
		Token:    officeToken,
		UserID:   userID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DebugSupabaseConfig returns debug info about Supabase configuration
func (h *AuthHandler) DebugSupabaseConfig(w http.ResponseWriter, r *http.Request) {
	projectRef := os.Getenv("SUPABASE_PROJECT_REF")
	jwksURL := ""
	if projectRef != "" {
		jwksURL = fmt.Sprintf("https://%s.supabase.co/auth/v1/.well-known/jwks.json", projectRef)
	}

	debug := map[string]interface{}{
		"supabase_enabled":   auth.IsSupabaseEnabled(),
		"has_project_ref":    projectRef != "",
		"project_ref":        projectRef,
		"project_ref_length": len(projectRef),
		"jwks_url":           jwksURL,
	}

	// Test JWKS URL if available
	if jwksURL != "" {
		resp, err := http.Get(jwksURL)
		if err != nil {
			debug["jwks_test_error"] = err.Error()
		} else {
			debug["jwks_test_status"] = resp.StatusCode
			resp.Body.Close()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debug)
}

// CheckUsernameAvailability checks if a username is available
func (h *AuthHandler) CheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate username format
	if len(req.Username) < 3 || len(req.Username) > 20 {
		response := map[string]interface{}{
			"available": false,
			"error":     "Username must be between 3 and 20 characters",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if username contains only allowed characters (alphanumeric and underscore)
	for _, char := range req.Username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			response := map[string]interface{}{
				"available": false,
				"error":     "Username can only contain letters, numbers, and underscores",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Check if username is already taken
	_, err := h.authService.GetUserByUsername(req.Username)
	if err == nil {
		// User found, username is taken
		response := map[string]interface{}{
			"available": false,
			"error":     "Username is already taken",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Username is available
	response := map[string]interface{}{
		"available": true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetUsername allows users to set their username (for Discord OAuth users)
func (h *AuthHandler) SetUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate username (same validation as availability check)
	if len(req.Username) < 3 || len(req.Username) > 20 {
		http.Error(w, "Username must be between 3 and 20 characters", http.StatusBadRequest)
		return
	}

	for _, char := range req.Username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			http.Error(w, "Username can only contain letters, numbers, and underscores", http.StatusBadRequest)
			return
		}
	}

	// Check if username is available
	_, err := h.authService.GetUserByUsername(req.Username)
	if err == nil {
		http.Error(w, "Username is already taken", http.StatusBadRequest)
		return
	}

	// Update the user's username
	err = h.authService.UpdateUsername(userID, req.Username)
	if err != nil {
		log.Printf("Error updating username for user %d: %v", userID, err)
		http.Error(w, "Error updating username", http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Username updated successfully",
	})
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
		return
	}

	// Refresh the token
	newToken, err := auth.RefreshToken(token)
	if err != nil {
		log.Printf("Token refresh error: %v", err)
		http.Error(w, "Token refresh failed", http.StatusUnauthorized)
		return
	}

	// Return new token
	resp := map[string]string{
		"token": newToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetVersion returns the current build version/timestamp
func (h *AuthHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	version := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"date":      time.Now().Format("2006-01-02 15:04:05 UTC"),
		"js_file":   "main.33c9e5bc.js",
		"build":     "latest",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}
