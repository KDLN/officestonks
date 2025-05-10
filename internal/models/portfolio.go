package models

// Portfolio represents a user's stock holding
type Portfolio struct {
	ID       int   `json:"id"`
	UserID   int   `json:"user_id"`
	StockID  int   `json:"stock_id"`
	Quantity int   `json:"quantity"`
	Stock    Stock `json:"stock,omitempty"` // For joined queries
}

// PortfolioRepository interface defines methods for portfolio data access
type PortfolioRepository interface {
	GetUserPortfolio(userID int) ([]*Portfolio, error)
	GetUserStockHolding(userID, stockID int) (*Portfolio, error)
	AddStockToPortfolio(userID, stockID, quantity int) error
	UpdateStockQuantity(portfolioID, newQuantity int) error
	RemoveStockFromPortfolio(portfolioID int) error
	CalculatePortfolioValue(userID int) (float64, error)
}

// PortfolioSummary provides an overview of a user's entire portfolio
type PortfolioSummary struct {
	CashBalance     float64      `json:"cash_balance"`
	StockValue      float64      `json:"stock_value"`
	TotalValue      float64      `json:"total_value"`
	PortfolioItems  []*Portfolio `json:"portfolio_items"`
}