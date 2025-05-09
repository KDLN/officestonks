package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"officestonks/internal/middleware"
	"officestonks/internal/models"
	"officestonks/internal/services"
)

// MarketHandler handles market-related requests
type MarketHandler struct {
	marketService *services.MarketService
}

// NewMarketHandler creates a new market handler
func NewMarketHandler(marketService *services.MarketService) *MarketHandler {
	return &MarketHandler{
		marketService: marketService,
	}
}

// GetAllStocks returns a list of all available stocks
func (h *MarketHandler) GetAllStocks(w http.ResponseWriter, r *http.Request) {
	stocks, err := h.marketService.GetAllStocks()
	if err != nil {
		http.Error(w, "Failed to retrieve stocks", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stocks)
}

// GetStockByID returns details for a specific stock
func (h *MarketHandler) GetStockByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stockID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}
	
	stock, err := h.marketService.GetStockByID(stockID)
	if err != nil {
		http.Error(w, "Stock not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stock)
}

// GetUserPortfolio returns the user's portfolio
func (h *MarketHandler) GetUserPortfolio(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	portfolio, err := h.marketService.GetUserPortfolio(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve portfolio", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(portfolio)
}

// TradeStock handles buy/sell requests
func (h *MarketHandler) TradeStock(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse request body
	var req models.TradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate input
	if req.StockID <= 0 || req.Quantity <= 0 {
		http.Error(w, "Invalid stock ID or quantity", http.StatusBadRequest)
		return
	}
	
	var err error
	
	// Execute the trade
	if req.Action == "buy" {
		err = h.marketService.BuyStock(userID, req.StockID, req.Quantity)
	} else if req.Action == "sell" {
		err = h.marketService.SellStock(userID, req.StockID, req.Quantity)
	} else {
		http.Error(w, "Invalid action, must be 'buy' or 'sell'", http.StatusBadRequest)
		return
	}
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Trade executed successfully",
	})
}

// GetTransactionHistory returns the user's transaction history
func (h *MarketHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse pagination parameters
	limit := 50
	offset := 0
	
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// Get the transactions
	transactions, err := h.marketService.GetUserTransactions(userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}