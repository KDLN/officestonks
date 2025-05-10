package repository

import (
	"database/sql"
	"errors"
	"time"

	"officestonks/internal/models"
)

// StockRepo implements the StockRepository interface
type StockRepo struct {
	db *sql.DB
}

// NewStockRepo creates a new stock repository
func NewStockRepo(db *sql.DB) *StockRepo {
	return &StockRepo{db: db}
}

// GetAllStocks retrieves all stocks from the database
func (r *StockRepo) GetAllStocks() ([]*models.Stock, error) {
	query := `
		SELECT id, symbol, name, sector, current_price, last_updated
		FROM stocks
		ORDER BY symbol ASC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stocks []*models.Stock
	for rows.Next() {
		var stock models.Stock
		err := rows.Scan(
			&stock.ID,
			&stock.Symbol,
			&stock.Name,
			&stock.Sector,
			&stock.CurrentPrice,
			&stock.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, &stock)
	}
	
	return stocks, nil
}

// GetStockByID retrieves a stock by ID
func (r *StockRepo) GetStockByID(id int) (*models.Stock, error) {
	var stock models.Stock
	
	query := `
		SELECT id, symbol, name, sector, current_price, last_updated
		FROM stocks
		WHERE id = ?
	`
	
	err := r.db.QueryRow(query, id).Scan(
		&stock.ID,
		&stock.Symbol,
		&stock.Name,
		&stock.Sector,
		&stock.CurrentPrice,
		&stock.LastUpdated,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("stock not found")
		}
		return nil, err
	}
	
	return &stock, nil
}

// GetStockBySymbol retrieves a stock by symbol
func (r *StockRepo) GetStockBySymbol(symbol string) (*models.Stock, error) {
	var stock models.Stock
	
	query := `
		SELECT id, symbol, name, sector, current_price, last_updated
		FROM stocks
		WHERE symbol = ?
	`
	
	err := r.db.QueryRow(query, symbol).Scan(
		&stock.ID,
		&stock.Symbol,
		&stock.Name,
		&stock.Sector,
		&stock.CurrentPrice,
		&stock.LastUpdated,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("stock not found")
		}
		return nil, err
	}
	
	return &stock, nil
}

// UpdateStockPrice updates a stock's price
func (r *StockRepo) UpdateStockPrice(stockID int, newPrice float64) error {
	query := `
		UPDATE stocks
		SET current_price = ?, last_updated = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, newPrice, time.Now(), stockID)
	return err
}

// LoadStocksForSimulation loads all stocks for the market simulator
func (r *StockRepo) LoadStocksForSimulation() (map[int]struct {
	ID       int
	Symbol   string
	Sector   string
	Price    float64
}, error) {
	query := `
		SELECT id, symbol, name, sector, current_price
		FROM stocks
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	stocks := make(map[int]struct {
		ID       int
		Symbol   string
		Sector   string
		Price    float64
	})
	
	for rows.Next() {
		var id int
		var symbol, name, sector string
		var price float64
		
		err := rows.Scan(&id, &symbol, &name, &sector, &price)
		if err != nil {
			return nil, err
		}
		
		stocks[id] = struct {
			ID       int
			Symbol   string
			Sector   string
			Price    float64
		}{
			ID:     id,
			Symbol: symbol,
			Sector: sector,
			Price:  price,
		}
	}
	
	return stocks, nil
}
// ResetAllStockPrices resets all stock prices to their initial values
func (r *StockRepo) ResetAllStockPrices() error {
	// This will reset all stock prices to their initial values in the database
	// Assuming initial prices are stored in a separate table or have a default value
	
	// The SQL statement to reset prices to original values
	// Using a transaction for atomicity
	tx, err := r.db.Begin()
	if err \!= nil {
		return err
	}
	
	// Option 1: Reset to the initial seed prices (assuming this is how your DB is set up)
	query := `
		UPDATE stocks
		SET current_price = CASE
			WHEN symbol = 'AAPL' THEN 150.00
			WHEN symbol = 'MSFT' THEN 250.00
			WHEN symbol = 'GOOGL' THEN 2500.00
			WHEN symbol = 'AMZN' THEN 3000.00
			WHEN symbol = 'META' THEN 300.00
			WHEN symbol = 'TSLA' THEN 700.00
			WHEN symbol = 'NFLX' THEN 550.00
			WHEN symbol = 'NVDA' THEN 600.00
			WHEN symbol = 'CRM' THEN 250.00
			WHEN symbol = 'PYPL' THEN 280.00
			ELSE current_price
		END,
		last_updated = ?
	`
	
	// Execute the update
	_, err = tx.Exec(query, time.Now())
	if err \!= nil {
		tx.Rollback()
		return err
	}
	
	// Commit the transaction
	return tx.Commit()
}
