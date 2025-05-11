package main

import (
	"log"
	"net/http"
)

// CORS middleware that allows all origins, methods, and headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the origin from the request
		origin := r.Header.Get("Origin")

		// Explicitly check for the frontend domain
		if origin == "https://officestonks-frontend-production.up.railway.app" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			log.Printf("CORS: Allowed specific frontend origin: %s", origin)
		} else if origin != "" {
			// Allow any other origin that provides Origin header
			w.Header().Set("Access-Control-Allow-Origin", origin)
			log.Printf("CORS: Allowed origin: %s", origin)
		} else {
			// Fall back to wildcard for requests without Origin
			w.Header().Set("Access-Control-Allow-Origin", "*")
			log.Printf("CORS: No origin header, using wildcard")
		}

		// Set standard CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Log request details for CORS debugging
		log.Printf("CORS middleware: Method=%s Path=%s Origin=%s",
			r.Method, r.URL.Path, origin)

		// Handle OPTIONS requests immediately
		if r.Method == "OPTIONS" {
			log.Printf("CORS: Responding to OPTIONS preflight request for %s", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process the actual request
		next.ServeHTTP(w, r)
	})
}