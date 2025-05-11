package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// RegisterDebugHandlers adds debug endpoints to the router
func RegisterDebugHandlers(r *http.ServeMux, repo interface{}) {
	// Convert repo to userRepo
	userRepo, ok := repo.(userRepository)
	if !ok {
		log.Println("WARNING: Could not register debug handlers - invalid repo type")
		return
	}

	// Ultra simple debug health check
	r.HandleFunc("/debug-health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte("Debug OK"))
	})

	// Debug handler for user list
	r.HandleFunc("/debug-users", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: /debug-users accessed from %s", r.Header.Get("Origin"))

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check method
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get users
		users, err := userRepo.GetAllUsers()
		if err != nil {
			log.Printf("DEBUG: Error getting users: %v", err)
			http.Error(w, fmt.Sprintf("Error getting users: %v", err), http.StatusInternalServerError)
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Debug users endpoint",
			"user_count": len(users),
			"users": users,
		})
	})

	// Debug handler for admin status
	r.HandleFunc("/debug-admin", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: /debug-admin accessed from %s", r.Header.Get("Origin"))

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check method
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user ID from query param
		userIDStr := r.URL.Query().Get("user_id")
		userID := 3 // Default to KDLN
		if userIDStr != "" {
			var err error
			userID, err = strconv.Atoi(userIDStr)
			if err != nil {
				log.Printf("DEBUG: Invalid user ID: %s", userIDStr)
			}
		}

		// Check admin status
		isAdmin, err := userRepo.IsUserAdmin(userID)
		if err != nil {
			log.Printf("DEBUG: Error checking admin status: %v", err)
			http.Error(w, fmt.Sprintf("Error checking admin status: %v", err), http.StatusInternalServerError)
			return
		}

		// Get detailed debug info
		debugInfo := ""
		if debugMethod, ok := userRepo.(interface{ DebugIsUserAdmin(int) string }); ok {
			debugInfo = debugMethod.DebugIsUserAdmin(userID)
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Debug admin endpoint",
			"user_id": userID,
			"is_admin": isAdmin,
			"debug_info": debugInfo,
		})
	})
}

// Simple interface for user repository methods we need
type userRepository interface {
	GetAllUsers() ([]*struct {
		ID           int    `json:"id"`
		Username     string `json:"username"`
		CashBalance  float64 `json:"cash_balance"`
		IsAdmin      bool   `json:"is_admin"`
	}, error)
	IsUserAdmin(userID int) (bool, error)
}