package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yourusername/officestonks/internal/auth"
	"github.com/yourusername/officestonks/internal/handlers"
	"github.com/yourusername/officestonks/internal/models"
	"github.com/yourusername/officestonks/internal/repository"
	"github.com/yourusername/officestonks/internal/services"

	"github.com/gorilla/mux"
	_ "github.com/go-sql-driver/mysql"
)

// TestDB is a test database connection
var TestDB *sql.DB

// SetupTestDB initializes a test database connection
func SetupTestDB() (*sql.DB, error) {
	// Use environment variables or hardcoded test database credentials
	username := getEnv("TEST_DB_USER", "root")
	password := getEnv("TEST_DB_PASSWORD", "password")
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "3306")
	dbname := getEnv("TEST_DB_NAME", "officestonks_test")

	// Create connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", 
		username, password, host, port, dbname)

	// Open connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// SetupTestRouter creates a router with all the handlers for testing
func SetupTestRouter(db *sql.DB) *mux.Router {
	// Create repositories
	userRepo := repository.NewUserRepo(db)
	stockRepo := repository.NewStockRepo(db)
	portfolioRepo := repository.NewPortfolioRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)

	// Create services
	authService := services.NewAuthService(userRepo)
	marketService := services.NewMarketService(stockRepo, userRepo, portfolioRepo, transactionRepo)

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService)
	marketHandler := handlers.NewMarketHandler(marketService)

	// Create middleware
	// authMiddleware := middleware.NewAuthMiddleware(authService)

	// Initialize router
	r := mux.NewRouter()

	// Set up API routes for testing
	apiRouter := r.PathPrefix("/api").Subrouter()
	
	// Auth routes
	authRouter := apiRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Stock routes
	apiRouter.HandleFunc("/stocks", marketHandler.GetAllStocks).Methods("GET")
	apiRouter.HandleFunc("/stocks/{id}", marketHandler.GetStockByID).Methods("GET")

	// Protected routes would be tested separately with auth token

	return r
}

// MakeRequest makes a test request and returns the response
func MakeRequest(method, url string, body interface{}, router *mux.Router) *httptest.ResponseRecorder {
	// Create request body if provided
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	// Create request
	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	rr := httptest.NewRecorder()
	
	// Serve request
	router.ServeHTTP(rr, req)
	
	return rr
}

// AuthenticatedRequest makes a test request with auth token
func AuthenticatedRequest(method, url string, body interface{}, userID int, router *mux.Router) *httptest.ResponseRecorder {
	// Create request body if provided
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	// Create request
	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	
	// Add auth token
	token, _ := auth.GenerateToken(userID)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// Record response
	rr := httptest.NewRecorder()
	
	// Serve request
	router.ServeHTTP(rr, req)
	
	return rr
}

// CreateTestUser creates a test user and returns the user model
func CreateTestUser(t *testing.T, router *mux.Router, username, password string) *models.AuthResponse {
	// Register user
	reqBody := models.AuthRequest{
		Username: username,
		Password: password,
	}
	
	rr := MakeRequest("POST", "/api/auth/register", reqBody, router)
	
	// Check status code
	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, rr.Code)
	}
	
	// Parse response
	var resp models.AuthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}
	
	return &resp
}