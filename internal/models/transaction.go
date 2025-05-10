package models

import (
	"time"
)

// TransactionType defines the type of transaction
type TransactionType string

const (
	Buy  TransactionType = "buy"
	Sell TransactionType = "sell"
)

// Transaction represents a stock purchase or sale
type Transaction struct {
	ID              int             `json:"id"`
	UserID          int             `json:"user_id"`
	StockID         int             `json:"stock_id"`
	Quantity        int             `json:"quantity"`
	Price           float64         `json:"price"`
	TransactionType TransactionType `json:"transaction_type"`
	CreatedAt       time.Time       `json:"created_at"`
	
	// For joined queries
	Stock           Stock           `json:"stock,omitempty"`
}

// TransactionRepository interface defines methods for transaction data access
type TransactionRepository interface {
	CreateTransaction(userID, stockID, quantity int, price float64, transType TransactionType) (*Transaction, error)
	GetUserTransactions(userID int, limit, offset int) ([]*Transaction, error)
}

// TradeRequest represents a buy or sell request
type TradeRequest struct {
	StockID  int    `json:"stock_id"`
	Quantity int    `json:"quantity"`
	Action   string `json:"action"` // "buy" or "sell"
}