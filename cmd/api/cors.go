package main

import (
	"log"
	"net/http"
	"strings"
)

// CORS middleware to allow frontend to communicate with the API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always log every request for debugging
		log.Printf("CORS Request: Method=%s Path=%s Origin=%s Host=%s",
			r.Method, r.URL.Path, r.Header.Get("Origin"), r.Host)

		// Get the origin from the request
		origin := r.Header.Get("Origin")

		// CRITICAL FIX: Always allow the production frontend origin unconditionally for all requests
		if origin == "https://officestonks-frontend-production.up.railway.app" {
			log.Printf("CORS: Explicitly allowing production frontend origin")
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			w.Header().Set("Vary", "Origin")
		} else if origin != "" {
			// Accept any other origin
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			w.Header().Set("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")
			log.Printf("CORS: Allowed origin: %s", origin)
		} else {
			// Fallback for no origin (like local testing with curl)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			log.Printf("CORS: No origin specified, using wildcard")
		}

		// Special explicit handling for admin routes
		if strings.Contains(r.URL.Path, "/api/admin/") {
			log.Printf("CORS: Admin endpoint detected: %s", r.URL.Path)
			// Be EXTRA sure CORS headers are set for admin
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			log.Printf("CORS: Responding to OPTIONS preflight request from %s", r.Header.Get("Origin"))
			// Return 200 OK for preflight
			w.WriteHeader(http.StatusOK)
			return
		}

		// Log all headers we're setting
		log.Printf("CORS Response Headers: %v", w.Header())

		// Process the actual request
		next.ServeHTTP(w, r)
	})
}