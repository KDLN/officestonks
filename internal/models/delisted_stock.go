package models

import "time"

// DelistingReason represents why a stock was delisted
type DelistingReason string

const (
	DelistingBankruptcy DelistingReason = "bankruptcy" // Company went bankrupt
	DelistingMerger     DelistingReason = "merger"     // Company was acquired/merged
	DelistingAdmin      DelistingReason = "admin"      // Manually delisted by admin
)

// DelistedStock represents a stock that has been removed from trading
type DelistedStock struct {
	ID         int             `json:"id"`
	Symbol     string          `json:"symbol"`
	Name       string          `json:"name"`
	Sector     string          `json:"sector"`
	FinalPrice float64         `json:"final_price"`
	DelistedAt time.Time       `json:"delisted_at"`
	Reason     DelistingReason `json:"reason"`
}

// DelistedStockRepository defines data access methods for delisted stocks
type DelistedStockRepository interface {
	CreateDelistedStock(stockID int, symbol, name, sector string, finalPrice float64, reason DelistingReason) error
	GetDelistedStocks(limit, offset int) ([]*DelistedStock, error)
	GetDelistedStockByID(id int) (*DelistedStock, error)
	GetRecentBankruptcies(limit int) ([]*DelistedStock, error)
}