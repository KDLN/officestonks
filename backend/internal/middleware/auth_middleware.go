package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/yourusername/officestonks/internal/services"
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
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		
		// Check if the header has the "Bearer " prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}
		
		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
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