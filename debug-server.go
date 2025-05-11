package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Simple debug server that runs separately from the main server
// Run with: go run debug-server.go
// Access at: http://localhost:8081/debug

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Simple debug info endpoint
	mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Content-Type", "application/json")

		// Return basic info
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Debug server is running",
			"time": fmt.Sprintf("%v", time.Now()),
			"environment": os.Environ(),
		})
	})

	// Get port from environment or use default 8081
	port := 8081
	if portStr := os.Getenv("DEBUG_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Start the server
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Debug server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}