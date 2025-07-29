package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"officestonks/internal/auth"
	"officestonks/internal/models"
	"officestonks/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
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
		"supabase_enabled":    auth.IsSupabaseEnabled(),
		"has_project_ref":     projectRef != "",
		"project_ref":         projectRef,
		"project_ref_length":  len(projectRef),
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