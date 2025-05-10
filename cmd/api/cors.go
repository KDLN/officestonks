package main

import (
	"fmt"
	"log"
	"net/http"
)

// CORS middleware to allow frontend to communicate with the API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request for debugging
		log.Printf("CORS: %s %s from Origin: %s", r.Method, r.URL.Path, r.Header.Get("Origin"))

		// Simply allow all origins for now to fix the immediate issue
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Add other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Debugging: Print out the headers we're setting
		fmt.Printf("Setting CORS headers: %v\n", w.Header())

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			log.Printf("Responding to OPTIONS request from %s", r.Header.Get("Origin"))
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}