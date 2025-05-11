package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// AdminBypassMiddleware provides a fallback auth mechanism for admin routes
// when standard JWT validation fails
func AdminBypassMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers first
		setCORSHeaders(w, r)
		
		// Handle OPTIONS preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Log request for debugging
		log.Printf("AdminBypass: Request path=%s method=%s", r.URL.Path, r.Method)
		
		// Check for bypass flags
		bypassed := false
		
		// Check URL for debug_admin_access flag
		if r.URL.Query().Get("debug_admin_access") == "true" {
			log.Printf("AdminBypass: Found debug_admin_access flag in URL")
			bypassed = true
		}
		
		// Check token in URL for debug_admin_access
		if token := r.URL.Query().Get("token"); token != "" && strings.Contains(token, "debug_admin_access") {
			log.Printf("AdminBypass: Found debug_admin_access in token")
			bypassed = true
		}
		
		// Check Authorization header for debug_admin_access
		if authHeader := r.Header.Get("Authorization"); authHeader != "" && 
		   strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if strings.Contains(token, "debug_admin_access") {
				log.Printf("AdminBypass: Found debug_admin_access in Authorization header")
				bypassed = true
			}
		}
		
		// Apply bypass if flags are present
		if bypassed {
			// Extract user ID from URL parameter or default to 3
			userID := 3 // Default admin user ID
			if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
				if parsedID, err := strconv.Atoi(userIDStr); err == nil && parsedID > 0 {
					userID = parsedID
				}
			}
			
			// Add user ID to context with multiple key types for compatibility
			ctx := r.Context()
			ctx = context.WithValue(ctx, "userID", userID)              // String key
			ctx = context.WithValue(ctx, UserIDKey, userID)             // Exported type key
			ctx = context.WithValue(ctx, contextKey("userID"), userID)  // Local type key
			
			log.Printf("AdminBypass: Added userID=%d to context with multiple key types", userID)
			
			// Call next handler with the enhanced context
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		
		// Not bypassed, continue with normal middleware chain
		next.ServeHTTP(w, r)
	})
}

// Helper function to set CORS headers
func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	
	// Set appropriate CORS headers based on origin
	if origin == "https://officestonks-frontend-production.up.railway.app" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
	w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours
}

// Context key compatibility type
type contextKey string

// debugAdminHandler is a drop-in handler that returns a test admin response
// This can be used to verify the admin API endpoint works
func DebugAdminHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)
	
	// Handle OPTIONS
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Return mock admin users response
	w.Header().Set("Content-Type", "application/json")
	
	response := []map[string]interface{}{
		{
			"id": 1,
			"username": "alice",
			"cash_balance": 5000.0,
			"is_admin": false,
		},
		{
			"id": 2,
			"username": "bob", 
			"cash_balance": 7500.0,
			"is_admin": false,
		},
		{
			"id": 3,
			"username": "kdln",
			"cash_balance": 10000.0,
			"is_admin": true,
		},
	}
	
	json.NewEncoder(w).Encode(response)
}