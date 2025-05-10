package main

import (
	"log"
	"net/http"
)

// CORS middleware that allows all origins, methods, and headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request for debugging
		log.Printf("CORS Request: Method=%s Path=%s Origin=%s Host=%s",
			r.Method, r.URL.Path, r.Header.Get("Origin"), r.Host)

		// Set wide-open CORS headers - allow everything
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Log the headers we're setting
		log.Printf("CORS: Set unrestricted headers for all origins")

		// Handle OPTIONS immediately
		if r.Method == "OPTIONS" {
			log.Printf("CORS: Handling OPTIONS request for %s", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}