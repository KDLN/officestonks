package main

import (
	"log"
	"net/http"
)

// CORS middleware to allow frontend to communicate with the API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// More detailed logging for CORS debugging
		log.Printf("CORS Request: Method=%s Path=%s Origin=%s Host=%s",
			r.Method, r.URL.Path, r.Header.Get("Origin"), r.Host)

		// Get the origin from the request
		origin := r.Header.Get("Origin")

		// Explicitly handle known origins
		knownOrigins := []string{
			"https://officestonks-frontend.vercel.app",
			"http://localhost:3000",
		}

		// Always set CORS headers regardless of origin
		if origin != "" {
			// Accept any origin to solve the immediate problem
			w.Header().Set("Access-Control-Allow-Origin", origin)

			// Log whether it's a known origin
			isKnown := false
			for _, known := range knownOrigins {
				if origin == known {
					isKnown = true
					break
				}
			}
			if isKnown {
				log.Printf("CORS: Allowed known origin: %s", origin)
			} else {
				log.Printf("CORS: Allowed unknown origin: %s", origin)
			}
		} else {
			// Fallback for no origin (like local testing with curl)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			log.Printf("CORS: No origin specified, using wildcard")
		}

		// Set all required CORS headers - be very explicit
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
		w.Header().Set("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")

		// Log all headers we're setting
		log.Printf("CORS Response Headers: %v", w.Header())

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			log.Printf("CORS: Responding to OPTIONS preflight request from %s", r.Header.Get("Origin"))
			// Return 200 OK for preflight
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process the actual request
		next.ServeHTTP(w, r)
	})
}