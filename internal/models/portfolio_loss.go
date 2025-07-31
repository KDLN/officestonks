package models

import "time"

// PortfolioLoss represents a loss from a stock bankruptcy
type PortfolioLoss struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	StockID     int       `json:"stock_id"`
	StockSymbol string    `json:"stock_symbol"`
	StockName   string    `json:"stock_name"`
	Quantity    int       `json:"quantity"`
	LossAmount  float64   `json:"loss_amount"`
	DelistedAt  time.Time `json:"delisted_at"`
}

// PortfolioLossRepository defines data access methods for portfolio losses
type PortfolioLossRepository interface {
	CreatePortfolioLoss(userID, stockID int, stockSymbol, stockName string, quantity int, lossAmount float64) error
	GetUserLosses(userID int, limit, offset int) ([]*PortfolioLoss, error)
	GetTotalUserLosses(userID int) (float64, error)
	GetRecentLosses(limit int) ([]*PortfolioLoss, error)
	GetLossesByStock(stockID int, limit int) ([]*PortfolioLoss, error)
}