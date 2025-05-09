package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/yourusername/officestonks/internal/models"
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