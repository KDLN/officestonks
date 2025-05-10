package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// setupStaticFileServer sets up routes to serve the API
func setupStaticFileServer(router *mux.Router) {
	// Root handler for the API
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip API and WebSocket routes
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws") {
			return
		}

		// For frontend routes, return a simple message directing users to the API
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Office Stonks API is running. Access the API at /api endpoints."}`))
	}).Methods("GET")
}