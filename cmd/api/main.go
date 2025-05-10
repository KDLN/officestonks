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

	// Initialize database connection with retries
	var db *sql.DB
	var err error

	// Try to connect to the database with retries
	for i := 0; i < 5; i++ {
		log.Printf("Attempting database connection (attempt %d of 5)...", i+1)
		db, err = repository.InitDB()
		if err == nil {
			log.Println("Successfully connected to database!")
			break
		}
		log.Printf("Failed to connect to database: %v", err)
		if i < 4 {
			log.Printf("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}

	// If all retries failed, check if we're in dev mode
	if err != nil {
		if os.Getenv("OFFICESTONKS_DEV_MODE") == "true" {
			log.Println("DEV MODE: Starting without database connection")
			// Create a simple health endpoint and exit
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Office Stonks API is running in DEV mode without DB"))
			})
			port := getPort()
			log.Printf("Server starting on port %d (DEV MODE)...\n", port)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
			return
		} else {
			log.Fatalf("All database connection attempts failed: %v", err)
		}
	}
	defer db.Close()

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

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create rate limiter (100 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	// Initialize router with middleware
	r := mux.NewRouter()

	// IMPORTANT: Apply middleware at the top level
	r.Use(corsMiddleware)
	r.Use(rateLimiter.RateLimit)

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

	// WebSocket route
	r.HandleFunc("/ws", wsHandler.HandleConnection)

	// Health check endpoint
	apiRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API is running"))
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

	// Set up static file serving for frontend
	setupStaticFileServer(r)

	// Get port from environment variable or use default
	port := getPort()
	fmt.Printf("Server starting on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
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

// CORS middleware is now defined in cors.go