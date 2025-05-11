package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// Logger middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request details
		log.Printf("DEBUG: %s %s", r.Method, r.URL.Path)
		log.Printf("DEBUG: Headers: %v", r.Header)
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// AdminStatusHandler returns admin status
func adminStatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("AdminStatusHandler called with URL: %s", r.URL.String())
	
	// Get token from query or header
	var token string
	if tokenParam := r.URL.Query().Get("token"); tokenParam != "" {
		token = tokenParam
		log.Printf("Token from query param: %s...", truncateString(token, 20))
	} else if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			log.Printf("Token from Authorization header: %s...", truncateString(token, 20))
		} else {
			log.Printf("Invalid Authorization header format: %s", authHeader)
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
	} else {
		log.Println("No token provided")
		http.Error(w, "No token provided", http.StatusUnauthorized)
		return
	}

	// Parse the token (in a real app we'd validate it)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		log.Printf("Invalid token format: %s", token)
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}

	// Extract user ID
	log.Printf("Token parts: %d segments", len(parts))
	
	// Return success with admin status = true
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"isAdmin": true,
		"userId": 3,
		"debug": map[string]interface{}{
			"tokenProvided": len(token) > 0,
			"tokenParts": len(parts),
		},
	}
	json.NewEncoder(w).Encode(response)
}

// HealthHandler returns server health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "healthy",
		"message": "Debug admin server is running",
	}
	json.NewEncoder(w).Encode(response)
}

// CORS middleware to handle preflight requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Helper to truncate string
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting debug admin server...")
	
	// Create router
	mux := http.NewServeMux()
	
	// Define routes
	mux.HandleFunc("/api/admin/status", adminStatusHandler)
	mux.HandleFunc("/api/health", healthHandler)
	
	// Wrap with middleware
	handler := corsMiddleware(loggingMiddleware(mux))
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	
	// Start server
	serverAddr := ":" + port
	log.Printf("Server listening on %s", serverAddr)
	log.Printf("Test with: curl http://localhost:%s/api/admin/status?token=YOUR_TOKEN", port)
	
	err := http.ListenAndServe(serverAddr, handler)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}