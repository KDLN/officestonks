package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/dgrijalva/jwt-go"    // Used indirectly
	_ "github.com/go-sql-driver/mysql" // Used as database driver
	"github.com/gorilla/mux"

	"officestonks/internal/handlers"
	"officestonks/internal/middleware"
	"officestonks/internal/repository"
	"officestonks/internal/services"
	"officestonks/internal/socketio"
	"officestonks/internal/websocket"
)

// Temporarily disabled embed to fix build
// //go:embed frontend/build
// var frontendFiles embed.FS

func main() {
	// Print startup information
	log.Println("🚀 Starting Office Stonks API server v1.1.3 (FORCE DEPLOY - Infinity Fix + Changelog Modal)...")
	log.Printf("Working directory: %s\n", getMustString("pwd"))
	log.Printf("Available files: %s\n", getMustString("ls -la"))
	log.Println("🔧 This deployment includes: Infinity sanitization, WebSocket fixes, Changelog modal")
	log.Println("📅 Deployment timestamp:", time.Now().Format("2006-01-02 15:04:05 MST"))

	// Remove reference to environment variables for production security
	log.Println("Environment loaded")

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

	// If all attempts failed, log error but continue (allow health check to respond)
	if err != nil {
		log.Printf("Failed to connect to database: %v. Server will start but database operations will fail.", err)
		log.Printf("Please ensure database environment variables are set correctly.")
		// Continue with nil db - services will need to handle this gracefully
	}

	// Verify connection with a simple query (only if db is not nil)
	if db != nil {
		var version string
		if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
			log.Printf("Warning: Could not verify database connection with query: %v", err)
		} else {
			log.Printf("Database connection verified. MySQL version: %s", version)
		}
	}
	// Create repositories
	log.Println("Creating database repositories...")
	userRepo := repository.NewUserRepo(db)
	stockRepo := repository.NewStockRepo(db)
	portfolioRepo := repository.NewPortfolioRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)
	chatRepo := repository.NewChatRepo(db)
	newsRepo := repository.NewNewsRepo(db)
	sectorRepo := repository.NewSectorRepo(db)
	changelogRepo := repository.NewChangelogRepo(db)
	auditRepo := repository.NewAuditRepo(db)
	delistedStockRepo := repository.NewDelistedStockRepo(db)
	portfolioLossRepo := repository.NewPortfolioLossRepo(db)
	sessionRepo := repository.NewSessionRepo(db)
	activityRepo := repository.NewActivityRepo(db)
	metricsRepo := repository.NewMetricsRepo(db)
	log.Println("✅ Repositories created successfully")

	// Create services
	log.Println("Creating application services...")
	authService := services.NewAuthService(userRepo)
	newsService := services.NewNewsService(newsRepo)
	marketService := services.NewMarketService(stockRepo, userRepo, portfolioRepo, transactionRepo, sectorRepo, delistedStockRepo, portfolioLossRepo, newsService)
	userService := services.NewUserService(userRepo, portfolioRepo)
	monitoringService := services.NewMonitoringService(sessionRepo, activityRepo, metricsRepo, auditRepo)
	log.Println("✅ Services created successfully")

	// Create websocket hub and initiate market simulator
	log.Println("Creating WebSocket hub...")
	wsHub := websocket.NewHub(marketService.GetSimulatorUpdates())
	go wsHub.Run()
	log.Println("✅ WebSocket hub started")

	// Initialize the market simulator after setting up the hub
	log.Println("🎯 Initializing market simulator...")
	if err := marketService.InitializeSimulator(); err != nil {
		log.Printf("❌ Failed to initialize market simulator: %v", err)
		log.Printf("⚠️ Server will continue but market simulation will not work")
	} else {
		log.Println("✅ Market simulator initialized successfully")
		log.Println("📊 Market simulator should now be generating stock updates every 2 seconds")
	}

	// Create chat service with the websocket hub
	chatService := services.NewChatService(chatRepo, userRepo, wsHub)
	changelogService := services.NewChangelogService(changelogRepo)
	auditService := services.NewAuditService(auditRepo)

	// Create websocket handler
	wsHandler := websocket.NewWebSocketHandler(wsHub, authService, monitoringService)
	
	// Create Railway-compatible handler for environments that don't support WebSocket hijacking
	railwayHandler := websocket.NewRailwayCompatibleHandler(wsHub, wsHandler)
	wsHub.SetRailwayHandler(railwayHandler)

	// Create native Socket.IO handler
	log.Println("🚀 Creating native Socket.IO v4 handler...")
	socketIOHandler := socketio.NewSocketIOHandler(marketService.GetSimulatorUpdates(), authService, monitoringService)
	log.Println("✅ Socket.IO handler ready")

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService, auditService, monitoringService)
	marketHandler := handlers.NewMarketHandler(marketService, monitoringService)
	userHandler := handlers.NewUserHandler(userService)
	chatHandler := handlers.NewChatHandler(chatService)
	adminHandler := handlers.NewAdminHandler(userRepo, stockRepo, chatRepo, marketService)
	newsHandler := handlers.NewNewsHandler(newsService)
	changelogHandler := handlers.NewChangelogHandler(changelogService)
	auditHandler := handlers.NewAuditHandler(auditService)
	gameConfigHandler := handlers.NewGameConfigHandler()
	testHandler := handlers.NewTestHandler(db, marketService, userService, stockRepo, portfolioRepo, delistedStockRepo, portfolioLossRepo, newsRepo)
	portfolioTestHandler := handlers.NewPortfolioTestHandler(marketService, userService, stockRepo, userRepo, portfolioRepo, transactionRepo)
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService)
	monitoringTestHandler := handlers.NewMonitoringTestHandler(db)
	log.Println("🔗 Creating SSE handler for real-time stock updates...")
	sseHandler := handlers.NewSSEHandler(marketService)
	log.Println("✅ SSE handler created and listening for stock updates")

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, monitoringService)

	// Create rate limiter (100 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(500, time.Minute) // Increased from 100 to 500 requests per minute
	
	// Create trade frequency limiter (5 second cooldown, max 20 trades per hour per user)
	tradeLimiter := middleware.NewTradeLimiter(5, 20)

	// Initialize router with middleware
	r := mux.NewRouter()

	// Define CORS middleware directly here
	corsMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the origin from the request
			origin := r.Header.Get("Origin")

			// Define allowed origins (now including direct frontend URL without proxy)
			allowedOrigins := []string{
				"https://officestonks-frontend-production.up.railway.app",
				"https://beta.officestonks.com",
				"https://officestonks.com",
				"http://localhost:3000",
				"http://localhost:3001",
			}

			// Check if origin is allowed
			originAllowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					originAllowed = true
					break
				}
			}

			// If origin is not in allowed list, allow it dynamically for development
			if !originAllowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if !originAllowed {
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

	// Apply CORS middleware first (needed for all routes)
	r.Use(corsMw)

	// SSE endpoint - must be registered BEFORE rate limiting to allow real-time updates
	r.HandleFunc("/api/sse/stock-updates", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔗 SSE connection request from: %s", r.Header.Get("Origin"))
		log.Printf("🔍 SSE request details - Method: %s, URL: %s, User-Agent: %s", 
			r.Method, r.URL.String(), r.Header.Get("User-Agent"))
		
		// Add Railway-specific headers BEFORE calling handler - match test endpoint exactly
		w.Header().Set("Content-Type", "text/event-stream")   // CRITICAL: Set SSE content type first
		w.Header().Set("X-Accel-Buffering", "no")           // Disable nginx buffering
		w.Header().Set("Cache-Control", "no-cache, no-store") // Prevent caching
		w.Header().Set("Connection", "keep-alive")            // Keep connection alive
		w.Header().Set("Access-Control-Allow-Origin", "*")    // Allow all origins
		
		// Handle OPTIONS requests for CORS preflight
		if r.Method == "OPTIONS" {
			log.Printf("✅ SSE OPTIONS preflight handled")
			w.WriteHeader(http.StatusOK)
			return
		}
		
		log.Printf("📡 Passing SSE request to handler...")
		sseHandler.HandleStockUpdates(w, r)
	}).Methods("GET", "OPTIONS")
	
	// SSE test endpoint to verify Railway compatibility
	r.HandleFunc("/api/sse/test", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🧪 SSE test endpoint called from: %s", r.Header.Get("Origin"))
		
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache, no-store")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no")
		
		// Send immediate test message with Railway-compatible format
		fmt.Fprintf(w, "data: {\"type\": \"test\", \"message\": \"SSE test successful\", \"timestamp\": %d}\r\n\r\n", time.Now().Unix())
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		
		log.Printf("✅ SSE test message sent")
	}).Methods("GET", "OPTIONS")
	
	// Working SSE endpoint - copy of test endpoint that actually sends stock updates
	r.HandleFunc("/api/sse/working", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🚀 Working SSE endpoint called from: %s", r.Header.Get("Origin"))
		
		// Use exact same headers as working test endpoint
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache, no-store")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no")
		
		// Send immediate connection message
		fmt.Fprintf(w, "data: {\"type\": \"connection\", \"status\": \"connected\", \"timestamp\": %d}\r\n\r\n", time.Now().Unix())
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		
		// Send current stock prices
		stocks, err := marketService.GetAllStocks()
		if err == nil {
			for _, stock := range stocks {
				fmt.Fprintf(w, "data: {\"type\": \"stock_update\", \"stock_id\": %d, \"symbol\": \"%s\", \"price\": %.2f}\r\n\r\n", 
					stock.ID, stock.Symbol, stock.CurrentPrice)
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		
		log.Printf("✅ Working SSE initial data sent")
		
		// Keep connection alive and send updates
		heartbeatTicker := time.NewTicker(15 * time.Second)
		stockUpdateTicker := time.NewTicker(2 * time.Second)
		defer heartbeatTicker.Stop()
		defer stockUpdateTicker.Stop()
		
		for {
			select {
			case <-stockUpdateTicker.C:
				// Send current stock prices every 2 seconds
				stocks, err := marketService.GetAllStocks()
				if err == nil {
					for _, stock := range stocks {
						fmt.Fprintf(w, "data: {\"type\": \"stock_update\", \"stock_id\": %d, \"symbol\": \"%s\", \"price\": %.2f}\r\n\r\n", 
							stock.ID, stock.Symbol, stock.CurrentPrice)
					}
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
				}
				
			case <-heartbeatTicker.C:
				// Send heartbeat
				fmt.Fprintf(w, "data: {\"type\": \"heartbeat\", \"timestamp\": %d}\r\n\r\n", time.Now().Unix())
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				
			case <-r.Context().Done():
				return
			}
		}
	}).Methods("GET", "OPTIONS")
	
	// Alternative HTTP polling endpoint as fallback
	r.HandleFunc("/api/stock-updates/poll", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Stock polling request from: %s", r.Header.Get("Origin"))
		
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		
		// Get current stock prices
		stocks, err := marketService.GetAllStocks()
		if err != nil {
			http.Error(w, "Failed to get stocks", http.StatusInternalServerError)
			return
		}
		
		// Format as stock updates
		updates := make([]map[string]interface{}, len(stocks))
		for i, stock := range stocks {
			updates[i] = map[string]interface{}{
				"type":     "stock_update",
				"stock_id": stock.ID,
				"symbol":   stock.Symbol,
				"price":    stock.CurrentPrice,
			}
		}
		
		response := map[string]interface{}{
			"type":      "stock_updates",
			"timestamp": time.Now().Unix(),
			"updates":   updates,
		}
		
		json.NewEncoder(w).Encode(response)
	}).Methods("GET", "OPTIONS")

	// Apply rate limiting after SSE endpoint registration
	r.Use(rateLimiter.RateLimit)
	r.Use(monitoringService.CreateRequestTrackerMiddleware())
	
	// Test endpoint to verify rate limiter bypass logic
	r.HandleFunc("/api/test-polling", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		log.Printf("🧪 TEST: Polling test endpoint called with X-Request-Type: %s", r.Header.Get("X-Request-Type"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Polling test endpoint working",
			"timestamp": time.Now().Unix(),
			"headers": map[string]string{
				"X-Request-Type": r.Header.Get("X-Request-Type"),
				"User-Agent": r.Header.Get("User-Agent"),
			},
		})
	}).Methods("GET", "OPTIONS")

	// Simple health check for Railway (no dependencies)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET", "OPTIONS")

	// Register a duplicate health check endpoint at the root level that accepts GET
	// This should help for Railway healthchecks
	r.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte("OK"))
	}).Methods("GET", "OPTIONS")

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
	authRouter.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/supabase", authHandler.SupabaseAuth).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/debug/supabase", authHandler.DebugSupabaseConfig).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/check-username", authHandler.CheckUsernameAvailability).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/version", authHandler.GetVersion).Methods("GET", "OPTIONS")

	// Protected auth routes
	protectedAuthRouter := authRouter.PathPrefix("").Subrouter()
	protectedAuthRouter.Use(authMiddleware.Authenticate)
	protectedAuthRouter.HandleFunc("/set-username", authHandler.SetUsername).Methods("POST", "OPTIONS")

	// Public market routes
	apiRouter.HandleFunc("/stocks", marketHandler.GetAllStocks).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/stocks/{id}", marketHandler.GetStockByID).Methods("GET", "OPTIONS")

	// Public user routes
	apiRouter.HandleFunc("/users/leaderboard", userHandler.GetLeaderboard).Methods("GET", "OPTIONS")

	// Public changelog routes
	apiRouter.HandleFunc("/changelog", changelogHandler.GetPublicChangelog).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/changelog/{version}", changelogHandler.GetChangelogByVersion).Methods("GET", "OPTIONS")
	
	// Test route to verify changelog system
	apiRouter.HandleFunc("/changelog/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"message": "Changelog test endpoint working",
			"test_entries": []map[string]interface{}{
				{
					"version": "v1.2.0",
					"title": "Crisis & News System",
					"is_major": true,
				},
			},
		})
	}).Methods("GET", "OPTIONS")

	// Database verification endpoint
	apiRouter.HandleFunc("/debug/database", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		result := map[string]interface{}{
			"timestamp": time.Now(),
			"database_connection": false,
			"tables": []string{},
			"monitoring_tables": map[string]bool{},
		}
		
		if db != nil {
			// Test connection
			if err := db.Ping(); err == nil {
				result["database_connection"] = true
				
				// Get all tables
				rows, err := db.Query("SHOW TABLES")
				if err == nil {
					var tables []string
					for rows.Next() {
						var tableName string
						if err := rows.Scan(&tableName); err == nil {
							tables = append(tables, tableName)
						}
					}
					rows.Close()
					result["tables"] = tables
					
					// Check specific monitoring tables
					monitoringTables := []string{"user_sessions", "user_activity", "system_metrics"}
					for _, table := range monitoringTables {
						exists := false
						for _, existing := range tables {
							if existing == table {
								exists = true
								break
							}
						}
						result["monitoring_tables"].(map[string]bool)[table] = exists
					}
				}
			} else {
				result["connection_error"] = err.Error()
			}
		} else {
			result["error"] = "Database connection is nil"
		}
		
		json.NewEncoder(w).Encode(result)
	}).Methods("GET", "OPTIONS")

	// Protected routes
	protectedRouter := apiRouter.PathPrefix("").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)
	protectedRouter.Use(tradeLimiter.Middleware)

	// Protected market routes
	protectedRouter.HandleFunc("/portfolio", marketHandler.GetUserPortfolio).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/trading", marketHandler.TradeStock).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/transactions", marketHandler.GetTransactionHistory).Methods("GET", "OPTIONS")

	// News for authenticated users
	protectedRouter.HandleFunc("/news", newsHandler.GetActiveNews).Methods("GET", "OPTIONS")

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
	adminRouter.HandleFunc("/stocks", adminHandler.GetAllStocksDetailed).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/stocks", adminHandler.CreateStock).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/stocks/{id:[0-9]+}", adminHandler.UpdateStockAdmin).Methods("PUT", "OPTIONS")
	adminRouter.HandleFunc("/stocks/{id:[0-9]+}", adminHandler.DeleteStockAdmin).Methods("DELETE", "OPTIONS")
	adminRouter.HandleFunc("/stocks/reset", adminHandler.ResetStockPrices).Methods("GET", "POST", "OPTIONS")
	adminRouter.HandleFunc("/stocks/ipo", adminHandler.LaunchIPO).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/stocks/sector-event", adminHandler.TriggerSectorEvent).Methods("POST", "OPTIONS")

	// Admin chat management
	adminRouter.HandleFunc("/chat/clear", adminHandler.ClearAllChats).Methods("GET", "POST", "OPTIONS")

	// Admin news posting
	adminRouter.HandleFunc("/news", newsHandler.CreateNews).Methods("POST", "OPTIONS")

	// Admin game configuration
	adminRouter.HandleFunc("/game-config", gameConfigHandler.GetGameConfig).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/game-config", gameConfigHandler.UpdateGameConfig).Methods("PUT", "OPTIONS")
	adminRouter.HandleFunc("/game-config/reset", gameConfigHandler.ResetGameConfig).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/game-config/balanced", gameConfigHandler.LoadBalancedConfig).Methods("POST", "OPTIONS")

	// Admin changelog management
	adminRouter.HandleFunc("/changelog", changelogHandler.GetAllChangelog).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/changelog", changelogHandler.CreateChangelog).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/changelog/{id:[0-9]+}/visibility", changelogHandler.UpdateChangelogVisibility).Methods("PUT", "OPTIONS")
	adminRouter.HandleFunc("/changelog/{id:[0-9]+}", changelogHandler.DeleteChangelog).Methods("DELETE", "OPTIONS")

	// Audit log
	adminRouter.HandleFunc("/audit", auditHandler.GetRecentEvents).Methods("GET", "OPTIONS")

	// Crisis testing endpoints
	adminRouter.HandleFunc("/crisis/force", adminHandler.ForceCrisisEvent).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/crisis/bankruptcy", adminHandler.ForceBankruptcy).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/crisis/recovery", adminHandler.ForceRecovery).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/simulator/status", adminHandler.GetSimulatorStatus).Methods("GET", "OPTIONS")

	// Test orchestration endpoints
	adminRouter.HandleFunc("/tests/status", testHandler.GetTestStatus).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/tests/crisis", testHandler.RunCrisisTests).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/tests/portfolio", portfolioTestHandler.RunPortfolioTests).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/tests/sse", testHandler.RunSSETests).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/tests/stock-management", testHandler.RunStockManagementTests).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/tests/create-sectors", testHandler.CreateMissingSectors).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/tests/admin-stock-update", testHandler.TestStockAdminUpdate).Methods("POST", "OPTIONS")

	// Monitoring endpoints
	adminRouter.HandleFunc("/monitoring/dashboard", monitoringHandler.GetMonitoringDashboard).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/monitoring/metrics", monitoringHandler.GetSystemMetrics).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/monitoring/sessions", monitoringHandler.GetActiveSessions).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/monitoring/activity", monitoringHandler.GetRecentActivity).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/monitoring/user-activity", monitoringHandler.GetUserActivity).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/monitoring/user-sessions", monitoringHandler.GetUserSessions).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/monitoring/activity-range", monitoringHandler.GetActivityByTimeRange).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/monitoring/test", monitoringTestHandler.TestMonitoring).Methods("GET", "OPTIONS")

	// WebSocket/SSE route with Railway-compatible fallback
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Real-time connection request: %s %s from %s", r.Method, r.URL.Path, r.Header.Get("Origin"))

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

		// Use Railway-compatible handler that automatically chooses WebSocket or SSE
		railwayHandler.HandleRealTimeConnection(w, r)
	})

	// Socket.IO native routes
	log.Println("🔌 Setting up native Socket.IO v4 routes...")
	
	// Handle all Socket.IO requests with native Go implementation
	r.PathPrefix("/socket.io/").Handler(socketIOHandler)
	
	// Socket.IO session status endpoint
	r.HandleFunc("/debug/socketio/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		// Get session info from handler
		sessionInfo := socketIOHandler.GetSessionStats()
		
		json.NewEncoder(w).Encode(sessionInfo)
	}).Methods("GET", "OPTIONS")
	
	// Socket.IO Admin stats endpoint
	r.HandleFunc("/api/admin/socketio/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		stats := map[string]interface{}{
			"status": "healthy",
			"protocol": "Socket.IO v4 (Native Go)",
			"transport": []string{"websocket", "polling"},
			"message": "Native Go Socket.IO implementation",
		}
		json.NewEncoder(w).Encode(stats)
	}).Methods("GET", "OPTIONS")
	
	// Socket.IO Debug page
	r.HandleFunc("/debug/socketio", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Socket.IO Debug - Office Stonks</title>
    <style>
        body { font-family: monospace; margin: 20px; background: #1a1a1a; color: #00ff00; }
        .container { max-width: 1200px; margin: 0 auto; }
        .section { margin: 20px 0; padding: 10px; border: 1px solid #333; }
        .log { background: #000; padding: 10px; height: 300px; overflow-y: scroll; font-size: 12px; }
        button { background: #333; color: #00ff00; border: 1px solid #666; padding: 10px; margin: 5px; cursor: pointer; }
        button:hover { background: #555; }
        .status { padding: 5px; margin: 5px 0; }
        .success { color: #00ff00; }
        .error { color: #ff0000; }
        .warning { color: #ffaa00; }
    </style>
    <script src="https://cdn.socket.io/4.7.2/socket.io.min.js"></script>
</head>
<body>
    <div class="container">
        <h1>🔌 Socket.IO Debug Console</h1>
        
        <div class="section">
            <h2>Connection Status</h2>
            <div id="status" class="status">Disconnected</div>
            <button onclick="testConnection()">Test WebSocket Connection</button>
            <button onclick="testPolling()">Test Polling Connection</button>
            <button onclick="disconnect()">Disconnect</button>
        </div>
        
        <div class="section">
            <h2>Connection Info</h2>
            <div id="info">
                <div>Transport: <span id="transport">None</span></div>
                <div>Socket ID: <span id="socketId">None</span></div>
                <div>Connected: <span id="connected">false</span></div>
            </div>
        </div>
        
        <div class="section">
            <h2>Test Actions</h2>
            <button onclick="subscribeStocks()">Subscribe to Stocks</button>
            <button onclick="joinChat()">Join Chat</button>
            <button onclick="sendPing()">Send Ping</button>
            <button onclick="sendTestMessage()">Send Test Message</button>
        </div>
        
        <div class="section">
            <h2>Debug Log</h2>
            <button onclick="clearLog()">Clear Log</button>
            <div id="log" class="log"></div>
        </div>
    </div>
    
    <script>
        let socket = null;
        const logDiv = document.getElementById('log');
        const statusDiv = document.getElementById('status');
        
        function log(message, type = 'info') {
            const timestamp = new Date().toLocaleTimeString();
            const className = type === 'error' ? 'error' : type === 'success' ? 'success' : type === 'warning' ? 'warning' : '';
            logDiv.innerHTML += '<div class="' + className + '">[' + timestamp + '] ' + message + '</div>';
            logDiv.scrollTop = logDiv.scrollHeight;
            console.log('[Socket.IO Debug]', message);
        }
        
        function updateStatus(message, type = 'info') {
            statusDiv.textContent = message;
            statusDiv.className = 'status ' + type;
        }
        
        function updateInfo() {
            if (socket) {
                document.getElementById('transport').textContent = socket.io?.engine?.transport?.name || 'Unknown';
                document.getElementById('socketId').textContent = socket.id || 'None';
                document.getElementById('connected').textContent = socket.connected;
            }
        }
        
        function testConnection() {
            log('🔄 Testing WebSocket connection...', 'info');
            
            const token = 'test_token_123'; // For debugging
            socket = io(window.location.origin, {
                transports: ['websocket', 'polling'],
                auth: { token: token },
                query: { token: token }
            });
            
            socket.on('connect', () => {
                log('✅ Connected successfully!', 'success');
                updateStatus('Connected', 'success');
                updateInfo();
            });
            
            socket.on('connect_error', (error) => {
                log('❌ Connection error: ' + error.message, 'error');
                updateStatus('Connection Error', 'error');
            });
            
            socket.on('disconnect', (reason) => {
                log('⚠️ Disconnected: ' + reason, 'warning');
                updateStatus('Disconnected', 'warning');
                updateInfo();
            });
            
            socket.on('connected', (data) => {
                log('📡 Server confirmation: ' + JSON.stringify(data), 'success');
            });
            
            socket.on('stock_update', (data) => {
                log('📊 Stock update: ' + JSON.stringify(data), 'info');
            });
            
            socket.on('subscription_confirmed', (data) => {
                log('✅ Subscription confirmed: ' + JSON.stringify(data), 'success');
            });
            
            socket.on('pong', (data) => {
                log('🏓 Pong received: ' + JSON.stringify(data), 'info');
            });
        }
        
        function testPolling() {
            log('🔄 Testing polling connection only...', 'info');
            
            const token = 'test_token_123';
            socket = io(window.location.origin, {
                transports: ['polling'], // Force polling only
                forceNew: true, // Force new connection
                auth: { token: token },
                query: { token: token },
                upgrade: false, // Disable WebSocket upgrade
                rememberUpgrade: false,
                timeout: 20000
            });
            
            // Same event handlers as above
            socket.on('connect', () => {
                log('✅ Polling connected successfully!', 'success');
                updateStatus('Connected (Polling)', 'success');
                updateInfo();
            });
            
            socket.on('connect_error', (error) => {
                log('❌ Polling connection error: ' + error.message, 'error');
                updateStatus('Polling Error', 'error');
            });
            
            socket.on('disconnect', (reason) => {
                log('⚠️ Polling disconnected: ' + reason, 'warning');
                updateStatus('Disconnected', 'warning');
                updateInfo();
            });
            
            socket.on('connected', (data) => {
                log('📡 Polling server confirmation: ' + JSON.stringify(data), 'success');
            });
            
            socket.on('stock_update', (data) => {
                log('📊 Polling stock update: ' + JSON.stringify(data), 'info');
            });
            
            socket.on('subscription_confirmed', (data) => {
                log('✅ Polling subscription confirmed: ' + JSON.stringify(data), 'success');
            });
            
            socket.on('pong', (data) => {
                log('🏓 Polling pong received: ' + JSON.stringify(data), 'info');
            });
        }
        
        function disconnect() {
            if (socket) {
                socket.disconnect();
                socket = null;
                updateStatus('Disconnected', 'warning');
                updateInfo();
                log('🔌 Manual disconnect', 'info');
            }
        }
        
        function subscribeStocks() {
            if (socket && socket.connected) {
                socket.emit('subscribe_stocks');
                log('📊 Subscribing to stocks...', 'info');
            } else {
                log('❌ Not connected', 'error');
            }
        }
        
        function joinChat() {
            if (socket && socket.connected) {
                socket.emit('join_chat');
                log('💬 Joining chat...', 'info');
            } else {
                log('❌ Not connected', 'error');
            }
        }
        
        function sendPing() {
            if (socket && socket.connected) {
                socket.emit('ping', Date.now());
                log('🏓 Ping sent', 'info');
            } else {
                log('❌ Not connected', 'error');
            }
        }
        
        function sendTestMessage() {
            if (socket && socket.connected) {
                socket.emit('test_message', { message: 'Hello from debug console!', timestamp: Date.now() });
                log('📤 Test message sent', 'info');
            } else {
                log('❌ Not connected', 'error');
            }
        }
        
        function clearLog() {
            logDiv.innerHTML = '';
        }
        
        // Auto-update info every second
        setInterval(updateInfo, 1000);
        
        log('🚀 Socket.IO Debug Console loaded', 'success');
        log('Token: test_token_123 (hardcoded for debugging)', 'info');
        log('Click "Test WebSocket Connection" to start', 'info');
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}).Methods("GET")
	
	log.Println("✅ Native Socket.IO routes configured:")
	log.Println("   📡 /socket.io/ - Native Go Socket.IO v4 handler")
	log.Println("   📊 /api/admin/socketio/stats - Admin API")
	log.Println("   🔧 /debug/socketio - Debug console")

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
			"time":   time.Now().Format(time.RFC3339),
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

		// Basic health check
		response := map[string]interface{}{
			"status":  "healthy",
			"service": "OfficeStonks API",
			"version": "1.0",
			"time":    time.Now().Format(time.RFC3339),
		}

		// Add database status
		if db != nil {
			if err := db.Ping(); err != nil {
				response["database"] = "disconnected"
				response["database_error"] = err.Error()
			} else {
				response["database"] = "connected"
			}
		} else {
			response["database"] = "not_initialized"
		}

		// Return health status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET", "OPTIONS")

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

	// Emergency stock price reset endpoint (no auth required - for fixing infinity errors)
	apiRouter.HandleFunc("/emergency/reset-stocks", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Emergency stock price reset called")

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		// Reset stock prices
		stocks, err := stockRepo.GetAllStocks()
		if err != nil {
			log.Printf("Emergency reset failed to get stocks: %v", err)
			http.Error(w, "Failed to get stocks", http.StatusInternalServerError)
			return
		}

		log.Printf("Emergency reset: Found %d stocks to reset", len(stocks))

		for _, stock := range stocks {
			// Generate a new random price between $1 and $1000
			newPrice := 1.0 + (rand.Float64() * 999.0)

			err = stockRepo.UpdateStockPrice(stock.ID, newPrice)
			if err != nil {
				log.Printf("Emergency reset: Failed to update %s: %v", stock.Symbol, err)
				continue
			}

			log.Printf("Emergency reset: Updated %s to $%.2f", stock.Symbol, newPrice)
		}

		// Validate market simulator
		log.Println("Emergency reset: Validating market simulator...")
		marketService.ValidateSimulator()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"message":      "Emergency stock reset completed",
			"stocks_reset": len(stocks),
		})
	}).Methods("GET", "POST", "OPTIONS")

	// Serve static files from frontend build directory (fallback to filesystem)
	staticDir := "./frontend/build/"

	// Check if frontend build exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Printf("WARNING: Frontend build directory not found at %s", staticDir)
		log.Printf("Current working directory: %s", getMustString("pwd"))
		log.Printf("Directory contents: %s", getMustString("ls -la"))

		// Serve a simple message instead of 404
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip API routes
			if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws") || r.URL.Path == "/health" || r.URL.Path == "/health-check" {
				http.NotFound(w, r)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
				<html>
				<body>
					<h1>Office Stonks API Server</h1>
					<p>Frontend not found. The API is running at /api endpoints.</p>
					<p><a href="/api/health">Check API Health</a></p>
				</body>
				</html>
			`))
		})
	} else {
		log.Printf("Frontend build directory found at %s", staticDir)

		// Serve static files
		r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir+"static/"))))

		// Serve favicon and other root assets
		r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, staticDir+"favicon.ico")
		})
		r.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, staticDir+"manifest.json")
		})
		r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, staticDir+"robots.txt")
		})

		// Serve index.html for all non-API routes (SPA routing)
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip API routes
			if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws") || r.URL.Path == "/health" || r.URL.Path == "/health-check" {
				http.NotFound(w, r)
				return
			}

			// Serve index.html for all other routes (React Router will handle routing)
			http.ServeFile(w, r, staticDir+"index.html")
		})
	}

	// ABSOLUTELY MINIMAL HEALTH CHECK
	r.Methods("GET").Path("/health-check").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte("OK"))
	})

	// Get port from environment variable or use default
	port := getPort()
	log.Printf("=== Office Stonks Server Starting ===")
	log.Printf("Port: %d", port)
	log.Printf("Health check endpoint: /health")
	log.Printf("API endpoints: /api/*")
	fmt.Printf("Server starting on port %d...\n", port)

	// Force IPv4 binding only - this helps with Railway's routing
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", port), // Bind to IPv4 only
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Server binding to 0.0.0.0:%d (IPv4 only)...", port)
	log.Printf("Health check URL: http://0.0.0.0:%d/health", port)
	log.Printf("=== Server Ready to Accept Connections ===")
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
