package main

import (
	"log"
	"net/http"
)

// CORS middleware to allow frontend to communicate with the API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request for debugging
		log.Printf("CORS: %s %s from Origin: %s", r.Method, r.URL.Path, r.Header.Get("Origin"))
		
		// Get the origin from the request
		origin := r.Header.Get("Origin")
		
		// Allow the specific origin that made the request
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// Fallback for no origin (like local testing with curl)
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		// Add necessary CORS headers for preflight requests
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
		
		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			log.Printf("Responding to OPTIONS request from %s", r.Header.Get("Origin"))
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Process the actual request
		next.ServeHTTP(w, r)
	})
}