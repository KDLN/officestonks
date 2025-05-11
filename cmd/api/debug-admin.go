package main

import (
	"encoding/json"
	"log"
	"net/http"
	"officestonks/internal/auth"
	"strings"
)

// RegisterJWTDebugHandlers registers JWT debug handlers
func RegisterJWTDebugHandlers(mux *http.ServeMux, userRepo interface{}) {
	// Debug handler for admin JWT parsing
	mux.HandleFunc("/debug-admin-jwt", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Get token from URL parameter or Authorization header
		token := r.URL.Query().Get("token")
		if token == "" {
			// Try Authorization header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		
		// If no token provided
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "No token provided",
				"message": "Provide a token via ?token= query parameter or Authorization: Bearer header",
			})
			return
		}
		
		// Parse token without validation
		tokenPreview := token
		if len(token) > 20 {
			tokenPreview = token[:20] + "..."
		}
		log.Printf("Debug API: Parsing token: %s", tokenPreview)
		
		// Parse the token using the ParseUnverified method
		parser := &auth.Parser{}
		claims, err := parser.ParseWithoutValidation(token)
		
		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			log.Printf("Debug API: Error parsing token: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid token format",
				"message": err.Error(),
				"token_preview": tokenPreview,
			})
			return
		}
		
		// Log the claims
		log.Printf("Debug API: Successfully parsed token for user ID: %d", claims.UserID)
		
		// Return claims
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Token parsed successfully",
			"token_preview": tokenPreview,
			"claims": claims,
		})
	})
}