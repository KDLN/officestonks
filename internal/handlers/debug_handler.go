package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// RegisterDebugEndpoints adds debug endpoints for troubleshooting
func RegisterDebugEndpoints(router *http.ServeMux) {
	log.Println("Registering debug endpoints")
	
	// Debug info endpoint
	router.HandleFunc("/api/debug/info", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Get request info
		token := r.URL.Query().Get("token")
		if token == "" {
			// Try Authorization header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		
		// Parse token if available
		var tokenInfo map[string]interface{}
		if token != "" {
			parts := strings.Split(token, ".")
			if len(parts) == 3 {
				payload, err := base64.RawURLEncoding.DecodeString(parts[1])
				if err != nil {
					// Try with padding
					payload, err = base64.URLEncoding.DecodeString(parts[1] + "==")
					if err == nil {
						json.Unmarshal(payload, &tokenInfo)
					}
				} else {
					json.Unmarshal(payload, &tokenInfo)
				}
			}
		}
		
		// Return debug info
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"request": map[string]interface{}{
				"method": r.Method,
				"path": r.URL.Path,
				"query": r.URL.RawQuery,
				"remote_addr": r.RemoteAddr,
				"headers": map[string]string{
					"User-Agent": r.UserAgent(),
					"Content-Type": r.Header.Get("Content-Type"),
					"Authorization": maskToken(r.Header.Get("Authorization")),
				},
			},
			"token_info": tokenInfo,
			"debug_mode": true,
			"timestamp": fmt.Sprintf("%v", r.Context().Value("requestTime")),
		}
		
		json.NewEncoder(w).Encode(response)
	})
	
	// Debug JWT endpoint
	router.HandleFunc("/api/debug/jwt", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Get token
		token := r.URL.Query().Get("token")
		if token == "" {
			// Try Authorization header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		
		if token == "" {
			http.Error(w, "No token provided", http.StatusBadRequest)
			return
		}
		
		// Check for debug_admin_access
		containsDebugFlag := strings.Contains(token, "debug_admin_access")
		
		// Parse token
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			http.Error(w, "Invalid token format", http.StatusBadRequest)
			return
		}
		
		// Try to decode payload
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			// Try with padding
			payload, err = base64.URLEncoding.DecodeString(parts[1] + "==")
			if err != nil {
				http.Error(w, "Failed to decode token payload", http.StatusBadRequest)
				return
			}
		}
		
		// Parse payload
		var claims map[string]interface{}
		err = json.Unmarshal(payload, &claims)
		if err != nil {
			http.Error(w, "Failed to parse token claims", http.StatusBadRequest)
			return
		}
		
		// Return JWT info
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"token_format_valid": true,
			"debug_flag_present": containsDebugFlag,
			"claims": claims,
			"raw_payload": string(payload),
			"debug_mode": true,
		}
		
		json.NewEncoder(w).Encode(response)
	})
}

// Helper to mask Authorization token in logs
func maskToken(token string) string {
	if token == "" {
		return ""
	}
	
	if strings.HasPrefix(token, "Bearer ") {
		return "Bearer ****"
	}
	
	return "****"
}
