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
	log.Println("✅ Repositories created successfully")

	// Create services
	log.Println("Creating application services...")
	authService := services.NewAuthService(userRepo)
	newsService := services.NewNewsService(newsRepo)
	marketService := services.NewMarketService(stockRepo, userRepo, portfolioRepo, transactionRepo, sectorRepo, delistedStockRepo, portfolioLossRepo, newsService)
	userService := services.NewUserService(userRepo, portfolioRepo)
	log.Println("✅ Services created successfully")

	// Create websocket hub and initiate market simulator
	log.Println("Creating WebSocket hub...")
	wsHub := websocket.NewHub(marketService.GetSimulatorUpdates())
	go wsHub.Run()
	log.Println("✅ WebSocket hub started")

	// Initialize the market simulator after setting up the hub
	log.Println("Initializing market simulator...")
	if err := marketService.InitializeSimulator(); err != nil {
		log.Printf("❌ Failed to initialize market simulator: %v", err)
		log.Printf("Server will continue but market simulation will not work")
	} else {
		log.Println("✅ Market simulator initialized successfully")
	}

	// Create chat service with the websocket hub
	chatService := services.NewChatService(chatRepo, userRepo, wsHub)
	changelogService := services.NewChangelogService(changelogRepo)
	auditService := services.NewAuditService(auditRepo)

	// Create websocket handler
	wsHandler := websocket.NewWebSocketHandler(wsHub, authService)

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService, auditService)
	marketHandler := handlers.NewMarketHandler(marketService)
	userHandler := handlers.NewUserHandler(userService)
	chatHandler := handlers.NewChatHandler(chatService)
	adminHandler := handlers.NewAdminHandler(userRepo, stockRepo, chatRepo, marketService)
	newsHandler := handlers.NewNewsHandler(newsService)
	changelogHandler := handlers.NewChangelogHandler(changelogService)
	auditHandler := handlers.NewAuditHandler(auditService)
	gameConfigHandler := handlers.NewGameConfigHandler()
	testHandler := handlers.NewTestHandler(marketService, userService, stockRepo, portfolioRepo, delistedStockRepo, portfolioLossRepo, newsRepo)

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

			// Define allowed origins (now including direct frontend URL without proxy)
			allowedOrigins := []string{
				"https://officestonks-frontend-production.up.railway.app",
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

	// IMPORTANT: Apply middleware at the top level
	r.Use(corsMw)
	r.Use(rateLimiter.RateLimit)

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

	// Protected routes
	protectedRouter := apiRouter.PathPrefix("").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)

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
	adminRouter.HandleFunc("/stocks/reset", adminHandler.ResetStockPrices).Methods("GET", "POST", "OPTIONS")

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
