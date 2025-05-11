package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/dgrijalva/jwt-go"     // Used indirectly
	_ "github.com/go-sql-driver/mysql"  // Used as database driver

	"officestonks/internal/handlers"
	"officestonks/internal/middleware"
	"officestonks/internal/repository"
	"officestonks/internal/services"
	"officestonks/internal/websocket"
)

func main() {
	// Print startup information
	log.Println("Starting Office Stonks API server...")
	log.Printf("Working directory: %s\n", getMustString("pwd"))
	log.Printf("Available files: %s\n", getMustString("ls -la"))
	log.Printf("Environment variables: %s\n", os.Environ())

	// Initialize database connection with retries and fallback
	var db *sql.DB
	var err error

	// Try to connect to the database with retries using environment variables
	for i := 0; i < 3; i++ {
		log.Printf("Attempting database connection with environment variables (attempt %d of 3)...", i+1)
		db, err = repository.InitDB()
		if err == nil {
			log.Println("Successfully connected to database using environment variables!")
			break
		}
		log.Printf("Failed to connect to database using environment variables: %v", err)
		if i < 2 {
			log.Printf("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}

	// If environment variable connection failed, try hardcoded parameters
	if err != nil {
		log.Println("All attempts to connect using environment variables failed. Trying hardcoded connection...")
		db, err = repository.InitDBHardcoded()
		if err != nil {
			log.Println("Hardcoded connection also failed. Trying alternative connection methods...")
			db, err = repository.TryAlternativeConnections()
			if err != nil {
				log.Fatalf("All database connection attempts failed: %v", err)
			}
			log.Println("Successfully connected to database using alternative connection methods!")
		} else {
			log.Println("Successfully connected to database using hardcoded parameters!")
		}
	}

	// Verify connection with a simple query
	var version string
	if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		log.Printf("Warning: Could not verify database connection with query: %v", err)
	} else {
		log.Printf("Database connection verified. MySQL version: %s", version)
	}
	// Create repositories
	userRepo := repository.NewUserRepo(db)
	stockRepo := repository.NewStockRepo(db)
	portfolioRepo := repository.NewPortfolioRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)
	chatRepo := repository.NewChatRepo(db)

	// Create services
	authService := services.NewAuthService(userRepo)
	marketService := services.NewMarketService(stockRepo, userRepo, portfolioRepo, transactionRepo)
	userService := services.NewUserService(userRepo, portfolioRepo)

	// Create websocket hub and initiate market simulator
	wsHub := websocket.NewHub(marketService.GetSimulatorUpdates())
	go wsHub.Run()

	// Initialize the market simulator after setting up the hub
	if err := marketService.InitializeSimulator(); err != nil {
		log.Fatalf("Failed to initialize market simulator: %v", err)
	}

	// Create chat service with the websocket hub
	chatService := services.NewChatService(chatRepo, userRepo, wsHub)

	// Create websocket handler
	wsHandler := websocket.NewWebSocketHandler(wsHub)

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService)
	marketHandler := handlers.NewMarketHandler(marketService)
	userHandler := handlers.NewUserHandler(userService)
	chatHandler := handlers.NewChatHandler(chatService)
	adminHandler := handlers.NewAdminHandler(userRepo, stockRepo, chatRepo)

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create rate limiter (100 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	// Initialize router with middleware
	r := mux.NewRouter()

	// Define CORS middleware directly here
	corsMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the origin from the request
			origin := r.Header.Get("Origin")
			
			// If origin is provided, use it; otherwise use wildcard
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			
			// Set standard CORS headers
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
			
			// Handle OPTIONS requests immediately
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			// Process the actual request
			next.ServeHTTP(w, r)
		})
	}
	
	// IMPORTANT: Apply middleware at the top level
	r.Use(corsMw)
	r.Use(rateLimiter.RateLimit)

	// Global OPTIONS handler to ensure CORS preflight works for all routes
	r.PathPrefix("/").Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Global OPTIONS handler for: %s", r.URL.Path)
		// CORS headers are set by the middleware, just return 200 OK
		w.WriteHeader(http.StatusOK)
	})

	// Set up API routes
	apiRouter := r.PathPrefix("/api").Subrouter()

	// Public routes
	authRouter := apiRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Public market routes
	apiRouter.HandleFunc("/stocks", marketHandler.GetAllStocks).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/stocks/{id}", marketHandler.GetStockByID).Methods("GET", "OPTIONS")

	// Public user routes
	apiRouter.HandleFunc("/users/leaderboard", userHandler.GetLeaderboard).Methods("GET", "OPTIONS")

	// Protected routes
	protectedRouter := apiRouter.PathPrefix("").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)

	// Protected market routes
	protectedRouter.HandleFunc("/portfolio", marketHandler.GetUserPortfolio).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/trading", marketHandler.TradeStock).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/transactions", marketHandler.GetTransactionHistory).Methods("GET", "OPTIONS")

	// Protected user routes
	protectedRouter.HandleFunc("/users/me", userHandler.GetUserProfile).Methods("GET", "OPTIONS")

	// Chat routes
	protectedRouter.HandleFunc("/chat/messages", chatHandler.GetRecentMessages).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/chat/send", chatHandler.SendMessage).Methods("POST", "OPTIONS")

	// Admin status check (for frontend)
	protectedRouter.HandleFunc("/admin/status", adminHandler.GetAdminStatus).Methods("GET", "OPTIONS")

	// Admin routes - protected by both auth middleware and admin check
	// Convert the AdminOnly middleware to a mux.MiddlewareFunc
	muxAdminMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			adminHandler.AdminOnly(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})(w, r)
		})
	}
	
	// Apply the converted middleware
	adminRouter := protectedRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(muxAdminMiddleware)

	// Admin user management
	adminRouter.HandleFunc("/users", adminHandler.GetAllUsers).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/users/{id:[0-9]+}", adminHandler.UpdateUser).Methods("PUT", "OPTIONS")
	adminRouter.HandleFunc("/users/{id:[0-9]+}", adminHandler.DeleteUser).Methods("DELETE", "OPTIONS")

	// Admin stock management
	adminRouter.HandleFunc("/stocks/reset", adminHandler.ResetStockPrices).Methods("GET", "POST", "OPTIONS")

	// Admin chat management
	adminRouter.HandleFunc("/chat/clear", adminHandler.ClearAllChats).Methods("GET", "POST", "OPTIONS")

	// WebSocket route with explicit OPTIONS handling
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("WebSocket request received: %s %s from %s", r.Method, r.URL.Path, r.Header.Get("Origin"))

		// Special handling for OPTIONS requests
		if r.Method == "OPTIONS" {
			// Set CORS headers
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(http.StatusOK)
			return
		}

		// For actual WebSocket connections, use the handler
		wsHandler.HandleConnection(w, r)
	})

	// WebSocket health check endpoint with proper CORS handling
	r.HandleFunc("/ws/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("WebSocket health check requested from: %s", r.Header.Get("Origin"))

		// Handle CORS properly
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Return health status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "WebSocket endpoint is available",
			"time": time.Now().Format(time.RFC3339),
			"server": "OfficeStonks Backend",
		})
	})

	// Health check endpoint with proper CORS handling
	apiRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API health check requested from: %s", r.Header.Get("Origin"))

		// Handle CORS properly
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Return health status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"service": "OfficeStonks API",
			"version": "1.0",
			"time": time.Now().Format(time.RFC3339),
		})
	}).Methods("GET", "OPTIONS")

	// TEMPORARY DEBUG API: Admin users without authentication
	apiRouter.HandleFunc("/debug/admin/users", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG Admin users endpoint accessed from: %s", r.Header.Get("Origin"))

		// Set CORS headers
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get users directly without authentication
		users, err := userRepo.GetAllUsers()
		if err != nil {
			log.Printf("DEBUG Admin: Error getting users: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal Server Error",
				"message": fmt.Sprintf("Error getting users: %v", err),
			})
			return
		}

		log.Printf("DEBUG Admin: Found %d users, returning them without auth", len(users))

		// Return users as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}).Methods("GET", "OPTIONS")

	// TEMPORARY DEBUG API: Admin status check without authentication
	apiRouter.HandleFunc("/debug/admin/status", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG Admin status endpoint accessed from: %s", r.Header.Get("Origin"))

		// Set CORS headers
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get user ID from query parameter (optional)
		userIDStr := r.URL.Query().Get("user_id")
		userID := 3 // Default to user 3 (KDLN)
		if userIDStr != "" {
			var err error
			userID, err = strconv.Atoi(userIDStr)
			if err != nil {
				log.Printf("Invalid user ID: %s", userIDStr)
				userID = 3
			}
		}

		// Check if user is admin
		isAdmin, err := userRepo.IsUserAdmin(userID)
		if err != nil {
			log.Printf("DEBUG Admin status: Error checking admin status: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Internal Server Error",
				"message": fmt.Sprintf("Error checking admin status: %v", err),
				"user_id": userID,
			})
			return
		}

		// Return debug info
		log.Printf("DEBUG Admin status: User %d isAdmin=%v", userID, isAdmin)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"user_id": userID,
			"is_admin": isAdmin,
			"timestamp": time.Now().Format(time.RFC3339),
			"debug_info": userRepo.DebugIsUserAdmin(userID),
		})
	}).Methods("GET", "OPTIONS")

	// CORS debug endpoint
	apiRouter.HandleFunc("/debug/cors", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS Debug Request: %s %s from %s", r.Method, r.URL.Path, r.Header.Get("Origin"))

		// Return information about the request headers for debugging
		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"success": true,
			"message": "CORS is working if you can see this response",
			"request_headers": r.Header,
			"origin": r.Header.Get("Origin"),
			"host": r.Host,
			"timestamp": time.Now().String(),
		}

		json.NewEncoder(w).Encode(response)
	}).Methods("GET", "POST", "OPTIONS")

	// Rate limiter statistics endpoint (admin only)
	apiRouter.HandleFunc("/stats/rate-limit", func(w http.ResponseWriter, r *http.Request) {
		// Check for admin token (simple implementation)
		token := r.URL.Query().Get("token")
		if token != os.Getenv("ADMIN_TOKEN") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		// Return rate limiter statistics
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rateLimiter.GetStats())
	}).Methods("GET", "OPTIONS")

	// Set up a simple handler for root path to return a welcome message
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Welcome to OfficeStonks API. Frontend is served separately.",
			"status": "running",
			"docs": "/api/health for health check",
		})
	})

	// ROOT LEVEL DEBUG ENDPOINTS FOR TESTING
	r.HandleFunc("/debug_admin_status", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("ROOT DEBUG ENDPOINT ACCESSED: /debug_admin_status from %s", r.Header.Get("Origin"))

		// Set CORS headers
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get user ID from query parameter (optional)
		userIDStr := r.URL.Query().Get("user_id")
		userID := 3 // Default to user 3 (KDLN)
		if userIDStr != "" {
			var err error
			userID, err = strconv.Atoi(userIDStr)
			if err != nil {
				log.Printf("Invalid user ID: %s", userIDStr)
				userID = 3
			}
		}

		// Check if user is admin
		isAdmin, err := userRepo.IsUserAdmin(userID)
		if err != nil {
			log.Printf("ROOT DEBUG: Error checking admin status: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Internal Server Error",
				"message": fmt.Sprintf("Error checking admin status: %v", err),
				"user_id": userID,
			})
			return
		}

		// Also get detailed debug info
		debugInfo := userRepo.DebugIsUserAdmin(userID)

		log.Printf("ROOT DEBUG: User %d isAdmin=%v, debug info: %s", userID, isAdmin, debugInfo)

		// Return users as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Debug admin status endpoint - NO AUTHENTICATION",
			"user_id": userID,
			"is_admin": isAdmin,
			"debug_info": debugInfo,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	r.HandleFunc("/debug_users", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("ROOT DEBUG ENDPOINT ACCESSED: /debug_users from %s", r.Header.Get("Origin"))

		// Set CORS headers
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Try to get users
		users, err := userRepo.GetAllUsers()
		if err != nil {
			log.Printf("ROOT DEBUG: Error getting users: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal Server Error",
				"message": fmt.Sprintf("Error getting users: %v", err),
			})
			return
		}

		log.Printf("ROOT DEBUG: Successfully retrieved %d users", len(users))

		// Return users as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Debug users endpoint - NO AUTHENTICATION",
			"user_count": len(users),
			"users": users,
		})
	})

	// Get port from environment variable or use default
	port := getPort()
	fmt.Printf("Server starting on port %d...\n", port)

	// Force IPv4 binding only - this helps with Railway's routing
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),  // Bind to IPv4 only
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("Server binding to 0.0.0.0:%d (IPv4 only)...\n", port)
	log.Fatal(server.ListenAndServe())
}

// Get port from environment or use default 8080
func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 8080
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 8080
	}

	return port
}

// Helper function to run commands and return output
func getMustString(command string) string {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command error: %s", err)
	}
	return strings.TrimSpace(string(output))
}