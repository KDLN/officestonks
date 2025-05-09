package main

import (
	"net/http"
)

// CORS middleware to allow frontend to communicate with the API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers with specific origins
		allowedOrigins := []string{
			"https://web-copy-production-5b48.up.railway.app",
			"http://localhost:3000",
			"http://localhost:8080",
		}
		
		// Log the origin for debugging
		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}
		
		// If no match, use wildcard (less secure but works for development)
		if w.Header().Get("Access-Control-Allow-Origin") == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		// Add other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}