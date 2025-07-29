package models

import (
	"time"
)

// StockStatus represents the current status of a stock
type StockStatus string

const (
	StockActive     StockStatus = "active"     // Normal trading
	StockDistressed StockStatus = "distressed" // At $0.01, at risk of delisting
	StockDelisted   StockStatus = "delisted"   // Removed from trading, holdings lost
)

// Stock represents a company stock in the market
type Stock struct {
	ID           int         `json:"id"`
	Symbol       string      `json:"symbol"`
	Name         string      `json:"name"`
	Sector       string      `json:"sector"`
	CurrentPrice float64     `json:"current_price"`
	Status       StockStatus `json:"status"`
	LastUpdated  time.Time   `json:"last_updated"`
}

// StockRepository interface defines methods for stock data access
type StockRepository interface {
	GetAllStocks() ([]*Stock, error)
	GetStockByID(id int) (*Stock, error)
	GetStockBySymbol(symbol string) (*Stock, error)
	UpdateStockPrice(stockID int, newPrice float64) error
	UpdateStockStatus(stockID int, status StockStatus) error
	GetDistressedStocks() ([]*Stock, error)
	DelistStock(stockID int) error
	LoadStocksForSimulation() (map[int]struct {
		ID       int
		Symbol   string
		Sector   string
		Price    float64
		Status   StockStatus
	}, error)
	ResetAllStockPrices() error
}

// StockPrice represents a simple price update
type StockPrice struct {
	StockID int     `json:"stock_id"`
	Symbol  string  `json:"symbol"`
	Price   float64 `json:"price"`
}