package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"officestonks/internal/services"
)

// Key type for context values
type ContextKey string

// UserIDKey is the context key for the user ID
const UserIDKey ContextKey = "userID"

// AuthMiddleware handles authentication for protected routes
type AuthMiddleware struct {
	authService *services.AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate verifies the JWT token and adds the user ID to the request context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always set CORS headers first, before any authentication checks
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

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		// Log request details for auth debugging
		log.Printf("Auth middleware: Method=%s Path=%s Origin=%s",
			r.Method, r.URL.Path, r.Header.Get("Origin"))

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			log.Printf("Auth middleware: Responding to OPTIONS preflight request")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check for token in query string
		tokenParam := r.URL.Query().Get("token")

		// Get the Authorization header if token not in query string
		authHeader := r.Header.Get("Authorization")
		var tokenString string

		if tokenParam != "" {
			// Use token from query parameter
			tokenString = tokenParam
			log.Printf("Auth middleware: Using token from URL parameter (length: %d)", len(tokenParam))
		} else if authHeader != "" {
			// Check if the header has the "Bearer " prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				log.Printf("Auth middleware: Invalid authorization format: %s", authHeader)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Unauthorized",
					"message": "Invalid authorization format",
				})
				return
			}

			// Extract the token from Authorization header
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			log.Printf("Auth middleware: Using token from Authorization header (length: %d)", len(tokenString))
		} else {
			// No token provided in either place
			log.Printf("Auth middleware: No token provided")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized",
				"message": "Authentication token required",
				"path": r.URL.Path,
			})
			return
		}

		// Use our bypass-enabled validation for tokens
		userID, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("Auth middleware: Token validation error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized",
				"message": "Invalid or expired token",
				"details": err.Error(),
			})
			return
		}
		log.Printf("Auth middleware: Valid token for user ID %d", userID)

		// Add the user ID to the request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the user ID from the request context
func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}