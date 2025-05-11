package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// CORS middleware that allows all origins
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("Starting static API server...")
	
	r := mux.NewRouter()
	r.Use(corsMiddleware)
	
	// API routes
	api := r.PathPrefix("/api").Subrouter()
	
	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"message": "Static API server is running",
		})
	}).Methods("GET", "OPTIONS")
	
	// Stocks endpoint
	api.HandleFunc("/stocks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		stocks := []map[string]interface{}{
			{"id": 1, "symbol": "AAPL", "name": "Apple Inc.", "sector": "Technology", "current_price": 150.25},
			{"id": 2, "symbol": "MSFT", "name": "Microsoft Corporation", "sector": "Technology", "current_price": 252.75},
			{"id": 3, "symbol": "GOOGL", "name": "Alphabet Inc.", "sector": "Technology", "current_price": 2530.50},
			{"id": 4, "symbol": "AMZN", "name": "Amazon.com Inc.", "sector": "Technology", "current_price": 3100.25},
			{"id": 5, "symbol": "META", "name": "Meta Platforms Inc.", "sector": "Technology", "current_price": 298.50},
			{"id": 6, "symbol": "TSLA", "name": "Tesla, Inc.", "sector": "Automotive", "current_price": 700.00},
			{"id": 7, "symbol": "NFLX", "name": "Netflix, Inc.", "sector": "Entertainment", "current_price": 550.00},
		}
		
		json.NewEncoder(w).Encode(stocks)
	}).Methods("GET", "OPTIONS")
	
	// Portfolio endpoint
	api.HandleFunc("/portfolio", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		portfolio := map[string]interface{}{
			"cash_balance": 10000.00,
			"holdings": []map[string]interface{}{
				{
					"stock_id": 1,
					"symbol": "AAPL",
					"name": "Apple Inc.",
					"quantity": 10,
					"current_price": 150.25,
					"value": 1502.50,
				},
				{
					"stock_id": 2,
					"symbol": "MSFT",
					"name": "Microsoft Corporation",
					"quantity": 5,
					"current_price": 252.75,
					"value": 1263.75,
				},
			},
			"total_value": 12766.25,
		}
		
		json.NewEncoder(w).Encode(portfolio)
	}).Methods("GET", "OPTIONS")
	
	// Transactions endpoint
	api.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		transactions := []map[string]interface{}{
			{
				"id": 1,
				"stock_id": 1,
				"symbol": "AAPL",
				"quantity": 10,
				"price": 145.25,
				"transaction_type": "buy",
				"created_at": time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
			},
			{
				"id": 2,
				"stock_id": 2,
				"symbol": "MSFT",
				"quantity": 5,
				"price": 240.50,
				"transaction_type": "buy",
				"created_at": time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
			},
		}
		
		json.NewEncoder(w).Encode(transactions)
	}).Methods("GET", "OPTIONS")
	
	// Admin endpoints
	admin := api.PathPrefix("/admin").Subrouter()
	
	// Admin status
	admin.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"isAdmin": true,
		})
	}).Methods("GET", "OPTIONS")
	
	// Admin users
	admin.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		users := []map[string]interface{}{
			{
				"id": 1,
				"username": "admin",
				"cash_balance": 10000.00,
				"is_admin": true,
				"created_at": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			},
			{
				"id": 2,
				"username": "user1",
				"cash_balance": 5000.00,
				"is_admin": false,
				"created_at": time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
			},
			{
				"id": 3,
				"username": "user2",
				"cash_balance": 7500.00,
				"is_admin": false,
				"created_at": time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
			},
		}
		
		json.NewEncoder(w).Encode(users)
	}).Methods("GET", "OPTIONS")
	
	// Admin reset stocks
	admin.HandleFunc("/stocks/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Stock prices reset successfully",
		})
	}).Methods("GET", "OPTIONS")
	
	// Admin clear chat
	admin.HandleFunc("/chat/clear", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Chat messages cleared successfully",
		})
	}).Methods("GET", "OPTIONS")
	
	// Chat endpoints
	api.HandleFunc("/chat/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		messages := []map[string]interface{}{
			{
				"id": 1,
				"user_id": 1,
				"username": "admin",
				"message": "Welcome to Office Stonks!",
				"created_at": time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
			},
			{
				"id": 2,
				"user_id": 2,
				"username": "user1",
				"message": "Thanks for the welcome!",
				"created_at": time.Now().Format(time.RFC3339),
			},
		}
		
		json.NewEncoder(w).Encode(messages)
	}).Methods("GET", "OPTIONS")
	
	// Chat send
	api.HandleFunc("/chat/send", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Parse the request body
		var message struct {
			Message string `json:"message"`
		}
		err := json.NewDecoder(r.Body).Decode(&message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Return mock response
		mockMessage := map[string]interface{}{
			"id": 3,
			"user_id": 1,
			"username": "admin",
			"message": message.Message,
			"created_at": time.Now().Format(time.RFC3339),
		}
		
		json.NewEncoder(w).Encode(mockMessage)
	}).Methods("POST", "OPTIONS")
	
	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Start server
	log.Printf("Static API server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}