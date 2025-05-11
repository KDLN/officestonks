package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"officestonks/internal/middleware"
	"officestonks/internal/repository"
)

// EMERGENCY DIRECT HANDLER - Bypasses all normal auth checks
func RegisterEmergencyAdminHandlers(router *http.ServeMux, userRepo repository.UserRepository) {
	log.Println("EMERGENCY: Registering direct admin handlers")
	
	// Direct admin users handler with no auth checks
	router.HandleFunc("/api/admin/users", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Log the request
		log.Printf("EMERGENCY ADMIN: Direct access to /api/admin/users from %s", r.RemoteAddr)
		
		// Get all users
		users, err := userRepo.GetAllUsers()
		if err != nil {
			log.Printf("Error getting users: %v", err)
			http.Error(w, "Error retrieving users", http.StatusInternalServerError)
			return
		}
		
		// Return users as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
			"debug_mode": true,
			"emergency_handler": true,
		})
	})
	
	// Direct admin status check - always returns true
	router.HandleFunc("/api/admin/status", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Log the request
		log.Printf("EMERGENCY ADMIN: Direct access to /api/admin/status from %s", r.RemoteAddr)
		
		// Return status as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isAdmin": true,
			"debug_mode": true,
			"emergency_handler": true,
		})
	})
}
