package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"officestonks/internal/handlers"
	"officestonks/internal/middleware"
	"officestonks/internal/repository"
	"officestonks/internal/services"
	"officestonks/internal/websocket"
)

func main() {
	// Initialize database connection
	db, err := repository.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create repositories
	userRepo := repository.NewUserRepo(db)
	stockRepo := repository.NewStockRepo(db)
	portfolioRepo := repository.NewPortfolioRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)

	// Create services
	authService := services.NewAuthService(userRepo)
	marketService := services.NewMarketService(stockRepo, userRepo, portfolioRepo, transactionRepo)

	// Initialize the market simulator
	if err := marketService.InitializeSimulator(); err != nil {
		log.Fatalf("Failed to initialize market simulator: %v", err)
	}

	// Create websocket handler
	wsHandler := websocket.NewWebSocketHandler(marketService.GetSimulatorUpdates())

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService)
	marketHandler := handlers.NewMarketHandler(marketService)

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Initialize router
	r := mux.NewRouter()

	// Set up API routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	
	// Public routes
	authRouter := apiRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Public market routes
	apiRouter.HandleFunc("/stocks", marketHandler.GetAllStocks).Methods("GET")
	apiRouter.HandleFunc("/stocks/{id}", marketHandler.GetStockByID).Methods("GET")

	// Protected routes
	protectedRouter := apiRouter.PathPrefix("").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)
	
	// Protected market routes
	protectedRouter.HandleFunc("/portfolio", marketHandler.GetUserPortfolio).Methods("GET")
	protectedRouter.HandleFunc("/trading", marketHandler.TradeStock).Methods("POST")
	protectedRouter.HandleFunc("/transactions", marketHandler.GetTransactionHistory).Methods("GET")

	// WebSocket route
	r.HandleFunc("/ws", wsHandler.HandleConnection)

	// Health check endpoint
	apiRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API is running"))
	}).Methods("GET")

	// Root endpoint
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Office Stonks API is running"))
	}).Methods("GET")

	// Set up CORS middleware (must be applied before routes)
	r = r.PathPrefix("").Subrouter()
	r.Use(corsMiddleware)

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

// CORS middleware to allow frontend to communicate with the API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers with specific origins
		allowedOrigins := []string{
			"https://web-copy-production-5b48.up.railway.app",
			"http://localhost:3000",
		}

		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// If no match, use wildcard (less secure but works for development)
		if w.Header().Get("Access-Control-Allow-Origin") == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Add other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}