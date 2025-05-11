package middleware

import (
	"context"
	"net/http"
	"strings"

	"officestonks/internal/services"
)

// Key type for context values
type contextKey string

// UserIDKey is the context key for the user ID
const UserIDKey contextKey = "userID"

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
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
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
		} else if authHeader != "" {
			// Check if the header has the "Bearer " prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			// Extract the token from Authorization header
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// No token provided in either place
			http.Error(w, "Authentication token required", http.StatusUnauthorized)
			return
		}

		// Validate the token
		userID, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

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