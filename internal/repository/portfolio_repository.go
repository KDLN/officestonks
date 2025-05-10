package repository

import (
	"database/sql"

	"officestonks/internal/models"
)

// PortfolioRepo implements the PortfolioRepository interface
type PortfolioRepo struct {
	db *sql.DB
}

// NewPortfolioRepo creates a new portfolio repository
func NewPortfolioRepo(db *sql.DB) *PortfolioRepo {
	return &PortfolioRepo{db: db}
}

// GetUserPortfolio gets all stocks in a user's portfolio
func (r *PortfolioRepo) GetUserPortfolio(userID int) ([]*models.Portfolio, error) {
	query := `
		SELECT p.id, p.user_id, p.stock_id, p.quantity,
			   s.symbol, s.name, s.sector, s.current_price
		FROM portfolios p
		JOIN stocks s ON p.stock_id = s.id
		WHERE p.user_id = ?
		ORDER BY s.symbol ASC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var portfolio []*models.Portfolio
	for rows.Next() {
		var p models.Portfolio
		var stock models.Stock
		
		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.StockID,
			&p.Quantity,
			&stock.Symbol,
			&stock.Name,
			&stock.Sector,
			&stock.CurrentPrice,
		)
		if err != nil {
			return nil, err
		}
		
		stock.ID = p.StockID
		p.Stock = stock
		portfolio = append(portfolio, &p)
	}
	
	return portfolio, nil
}

// GetUserStockHolding gets a specific stock holding for a user
func (r *PortfolioRepo) GetUserStockHolding(userID, stockID int) (*models.Portfolio, error) {
	var p models.Portfolio
	var stock models.Stock
	
	query := `
		SELECT p.id, p.user_id, p.stock_id, p.quantity,
			   s.symbol, s.name, s.sector, s.current_price
		FROM portfolios p
		JOIN stocks s ON p.stock_id = s.id
		WHERE p.user_id = ? AND p.stock_id = ?
	`
	
	err := r.db.QueryRow(query, userID, stockID).Scan(
		&p.ID,
		&p.UserID,
		&p.StockID,
		&p.Quantity,
		&stock.Symbol,
		&stock.Name,
		&stock.Sector,
		&stock.CurrentPrice,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No holding for this stock
		}
		return nil, err
	}
	
	stock.ID = p.StockID
	p.Stock = stock
	return &p, nil
}

// AddStockToPortfolio adds a stock to a user's portfolio
func (r *PortfolioRepo) AddStockToPortfolio(userID, stockID, quantity int) error {
	// Check if the user already has this stock
	existing, err := r.GetUserStockHolding(userID, stockID)
	if err != nil {
		return err
	}
	
	if existing != nil {
		// Update existing holding
		query := `
			UPDATE portfolios
			SET quantity = quantity + ?
			WHERE user_id = ? AND stock_id = ?
		`
		_, err := r.db.Exec(query, quantity, userID, stockID)
		return err
	} else {
		// Create new holding
		query := `
			INSERT INTO portfolios (user_id, stock_id, quantity)
			VALUES (?, ?, ?)
		`
		_, err := r.db.Exec(query, userID, stockID, quantity)
		return err
	}
}

// UpdateStockQuantity updates the quantity of a stock in a portfolio
func (r *PortfolioRepo) UpdateStockQuantity(portfolioID, newQuantity int) error {
	if newQuantity <= 0 {
		return r.RemoveStockFromPortfolio(portfolioID)
	}
	
	query := `
		UPDATE portfolios
		SET quantity = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, newQuantity, portfolioID)
	return err
}

// RemoveStockFromPortfolio removes a stock from a portfolio
func (r *PortfolioRepo) RemoveStockFromPortfolio(portfolioID int) error {
	query := `
		DELETE FROM portfolios
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, portfolioID)
	return err
}

// CalculatePortfolioValue calculates the total value of a user's portfolio
func (r *PortfolioRepo) CalculatePortfolioValue(userID int) (float64, error) {
	// First, get the user's cash balance
	var cashBalance float64
	cashQuery := `
		SELECT cash_balance
		FROM users
		WHERE id = ?
	`
	err := r.db.QueryRow(cashQuery, userID).Scan(&cashBalance)
	if err != nil {
		return 0, err
	}
	
	// Then, calculate the value of all stocks
	stockValueQuery := `
		SELECT COALESCE(SUM(p.quantity * s.current_price), 0) as stock_value
		FROM portfolios p
		JOIN stocks s ON p.stock_id = s.id
		WHERE p.user_id = ?
	`
	var stockValue float64
	err = r.db.QueryRow(stockValueQuery, userID).Scan(&stockValue)
	if err != nil {
		return 0, err
	}
	
	// Return total portfolio value
	return cashBalance + stockValue, nil
}