package models

import (
	"time"
)

// Stock represents a company stock in the market
type Stock struct {
	ID           int       `json:"id"`
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	Sector       string    `json:"sector"`
	CurrentPrice float64   `json:"current_price"`
	LastUpdated  time.Time `json:"last_updated"`
}

// StockRepository interface defines methods for stock data access
type StockRepository interface {
	GetAllStocks() ([]*Stock, error)
	GetStockByID(id int) (*Stock, error)
	GetStockBySymbol(symbol string) (*Stock, error)
	UpdateStockPrice(stockID int, newPrice float64) error
	LoadStocksForSimulation() (map[int]struct {
		ID       int
		Symbol   string
		Sector   string
		Price    float64
	}, error)
	ResetAllStockPrices() error
}

// StockPrice represents a simple price update
type StockPrice struct {
	StockID int     `json:"stock_id"`
	Symbol  string  `json:"symbol"`
	Price   float64 `json:"price"`
}