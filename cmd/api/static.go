package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// setupStaticFileServer sets up routes to serve the React frontend
func setupStaticFileServer(router *mux.Router) {
	// Serve static files from the frontend build directory
	staticDir := "./frontend/build/"
	
	// Check if build directory exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		// Fallback: serve from current directory build folder
		staticDir = "./build/"
	}

	// Handle static assets (CSS, JS, images)
	staticFileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir+"static/")))
	router.PathPrefix("/static/").Handler(staticFileServer)

	// Handle manifest.json, favicon.ico, etc.
	router.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "manifest.json"))
	})
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "favicon.ico"))
	})

	// Catch-all handler for React routes (SPA routing)
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip API, WebSocket, emergency, and debug routes
		if strings.HasPrefix(r.URL.Path, "/api/") ||
		   strings.HasPrefix(r.URL.Path, "/ws") ||
		   strings.HasPrefix(r.URL.Path, "/emergency/") ||
		   strings.HasPrefix(r.URL.Path, "/debug_") ||
		   strings.HasPrefix(r.URL.Path, "/health") {
			return
		}

		// For all other routes, serve the React app's index.html
		indexPath := filepath.Join(staticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
		} else {
			// Fallback message if no build exists
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Frontend build not found. Run 'npm run build' in the frontend directory."}`))
		}
	}).Methods("GET", "OPTIONS")
}