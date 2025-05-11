package main

import (
	"log"
	"net/http"
)

// CORS middleware that allows all origins, methods, and headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always set CORS headers at the beginning of every request
		origin := r.Header.Get("Origin")
		if origin == "" {
			// If no origin is provided, allow any origin
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Otherwise, echo back the specific origin
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Set all required CORS headers
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Log request for debugging
		log.Printf("CORS Request: Method=%s Path=%s Origin=%s Host=%s",
			r.Method, r.URL.Path, origin, r.Host)
		log.Printf("CORS: Set headers to allow origin: %s", w.Header().Get("Access-Control-Allow-Origin"))

		// Handle OPTIONS immediately
		if r.Method == "OPTIONS" {
			log.Printf("CORS: Handling OPTIONS preflight request for %s", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}