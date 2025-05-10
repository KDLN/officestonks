package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/gorilla/mux"
)

// setupStaticFileServer sets up routes to serve the frontend static files
func setupStaticFileServer(router *mux.Router) {
	// Define the directory where static files are located
	staticDir := "./frontend/build"
	
	// Check if the directory exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		fmt.Printf("Warning: Static files directory %s does not exist\n", staticDir)
		return
	}
	
	// Create a file server that serves files from the frontend/build directory
	fileServer := http.FileServer(http.Dir(staticDir))
	
	// Serve static files (JS, CSS, etc.) directly
	router.PathPrefix("/static/").Handler(fileServer)
	router.PathPrefix("/favicon.ico").Handler(fileServer)
	router.PathPrefix("/manifest.json").Handler(fileServer)
	router.PathPrefix("/robots.txt").Handler(fileServer)
	
	// For all other routes that aren't API or static files, serve index.html
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip API and WebSocket routes
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws") {
			return
		}
		
		// Check if the requested file exists
		path := filepath.Join(staticDir, r.URL.Path)
		_, err := os.Stat(path)
		
		// If the file exists and is not a directory, serve it directly
		if err == nil {
			fileInfo, _ := os.Stat(path)
			if !fileInfo.IsDir() {
				http.StripPrefix("/", fileServer).ServeHTTP(w, r)
				return
			}
		}
		
		// Otherwise serve index.html for client-side routing
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	}).Methods("GET")
}