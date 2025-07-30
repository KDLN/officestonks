package repository

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"
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
		SELECT id, symbol, name, sector, sector_id, current_price, status, 
		       crisis_start, recovery_chance, bankruptcy_chance, last_updated
		FROM stocks
		WHERE status != 'delisted'
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
		var statusStr string
		err := rows.Scan(
			&stock.ID,
			&stock.Symbol,
			&stock.Name,
			&stock.Sector,
			&stock.SectorID,
			&stock.CurrentPrice,
			&statusStr,
			&stock.CrisisStart,
			&stock.RecoveryChance,
			&stock.BankruptcyChance,
			&stock.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		stock.Status = models.StockStatus(statusStr)
		stocks = append(stocks, &stock)
	}
	
	return stocks, nil
}

// GetStockByID retrieves a stock by ID
func (r *StockRepo) GetStockByID(id int) (*models.Stock, error) {
	var stock models.Stock
	var statusStr string
	
	query := `
		SELECT id, symbol, name, sector, current_price, status, last_updated
		FROM stocks
		WHERE id = ?
	`
	
	err := r.db.QueryRow(query, id).Scan(
		&stock.ID,
		&stock.Symbol,
		&stock.Name,
		&stock.Sector,
		&stock.CurrentPrice,
		&statusStr,
		&stock.LastUpdated,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("stock not found")
		}
		return nil, err
	}
	
	stock.Status = models.StockStatus(statusStr)
	
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
	SectorID int
	Price    float64
	Status   models.StockStatus
}, error) {
	query := `
		SELECT id, symbol, name, sector, COALESCE(sector_id, 0), current_price, status
		FROM stocks
		WHERE status != 'delisted'
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
		SectorID int
		Price    float64
		Status   models.StockStatus
	})
	
	for rows.Next() {
		var id int
		var symbol, name, sector string
		var sectorID int
		var price float64
		var statusStr string
		
		err := rows.Scan(&id, &symbol, &name, &sector, &sectorID, &price, &statusStr)
		if err != nil {
			return nil, err
		}
		
		stocks[id] = struct {
			ID       int
			Symbol   string
			Sector   string
			SectorID int
			Price    float64
			Status   models.StockStatus
		}{
			ID:       id,
			Symbol:   symbol,
			Sector:   sector,
			SectorID: sectorID,
			Price:    price,
			Status:   models.StockStatus(statusStr),
		}
	}
	
	return stocks, nil
}

// ResetAllStockPrices resets all stock prices to random values
func (r *StockRepo) ResetAllStockPrices() error {
	log.Println("ResetAllStockPrices: Starting to reset stock prices")
	
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Get all stocks
	stocks, err := r.GetAllStocks()
	if err != nil {
		log.Printf("ResetAllStockPrices: Error getting stocks: %v", err)
		return err
	}
	
	log.Printf("ResetAllStockPrices: Found %d stocks to reset", len(stocks))
	
	// Update each stock individually with a random price
	for _, stock := range stocks {
		// Generate a random price between 50 and 1000
		newPrice := 50.0 + rand.Float64()*950.0
		
		// Round to 2 decimal places
		newPrice = float64(int(newPrice*100)) / 100
		
		log.Printf("ResetAllStockPrices: Updating %s price from %.2f to %.2f", 
			stock.Symbol, stock.CurrentPrice, newPrice)
		
		// Update the stock price
		updateQuery := `
			UPDATE stocks
			SET current_price = ?,
				last_updated = ?
			WHERE id = ?
		`
		
		_, err := r.db.Exec(updateQuery, newPrice, time.Now(), stock.ID)
		if err != nil {
			log.Printf("ResetAllStockPrices: Error updating stock %s: %v", stock.Symbol, err)
			return err
		}
	}
	
	log.Println("ResetAllStockPrices: Successfully reset all stock prices")
	return nil
}

// UpdateStockStatus updates the status of a stock
func (r *StockRepo) UpdateStockStatus(stockID int, status models.StockStatus) error {
	query := `
		UPDATE stocks 
		SET status = ?, last_updated = ? 
		WHERE id = ?
	`
	
	_, err := RetryExec(r.db, query, string(status), time.Now(), stockID)
	if err != nil {
		log.Printf("Error updating stock status for ID %d: %v", stockID, err)
		return err
	}
	
	log.Printf("Updated stock %d status to %s", stockID, status)
	return nil
}

// GetDistressedStocks returns all stocks with distressed status
func (r *StockRepo) GetDistressedStocks() ([]*models.Stock, error) {
	query := `
		SELECT id, symbol, name, sector, current_price, status, last_updated
		FROM stocks 
		WHERE status = 'distressed'
		ORDER BY symbol
	`
	
	rows, err := RetryQuery(r.db, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stocks []*models.Stock
	for rows.Next() {
		stock := &models.Stock{}
		var statusStr string
		err := rows.Scan(&stock.ID, &stock.Symbol, &stock.Name, &stock.Sector, 
			&stock.CurrentPrice, &statusStr, &stock.LastUpdated)
		if err != nil {
			return nil, err
		}
		stock.Status = models.StockStatus(statusStr)
		stocks = append(stocks, stock)
	}
	
	return stocks, nil
}

// DelistStock permanently removes a stock from trading (sets status to delisted)
func (r *StockRepo) DelistStock(stockID int) error {
	query := `
		UPDATE stocks 
		SET status = 'delisted', last_updated = ? 
		WHERE id = ?
	`
	
	_, err := RetryExec(r.db, query, time.Now(), stockID)
	if err != nil {
		log.Printf("Error delisting stock ID %d: %v", stockID, err)
		return err
	}
	
	log.Printf("Delisted stock ID %d", stockID)
	return nil
}