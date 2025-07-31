package repository

import (
	"database/sql"

	"officestonks/internal/models"
)

// DelistedStockRepo implements models.DelistedStockRepository backed by MySQL
type DelistedStockRepo struct {
	db *sql.DB
}

// NewDelistedStockRepo creates a new DelistedStockRepo
func NewDelistedStockRepo(db *sql.DB) *DelistedStockRepo {
	return &DelistedStockRepo{db: db}
}

// CreateDelistedStock records a stock delisting
func (r *DelistedStockRepo) CreateDelistedStock(stockID int, symbol, name, sector string, finalPrice float64, reason models.DelistingReason) error {
	query := `
		INSERT INTO delisted_stocks (id, symbol, name, sector, final_price, reason, delisted_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := RetryExec(r.db, query, stockID, symbol, name, sector, finalPrice, string(reason))
	return err
}

// GetDelistedStocks returns paginated delisted stocks
func (r *DelistedStockRepo) GetDelistedStocks(limit, offset int) ([]*models.DelistedStock, error) {
	query := `
		SELECT id, symbol, name, sector, final_price, delisted_at, reason
		FROM delisted_stocks
		ORDER BY delisted_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := RetryQuery(r.db, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []*models.DelistedStock
	for rows.Next() {
		var stock models.DelistedStock
		var reasonStr string
		err := rows.Scan(
			&stock.ID,
			&stock.Symbol,
			&stock.Name,
			&stock.Sector,
			&stock.FinalPrice,
			&stock.DelistedAt,
			&reasonStr,
		)
		if err != nil {
			return nil, err
		}
		stock.Reason = models.DelistingReason(reasonStr)
		stocks = append(stocks, &stock)
	}
	return stocks, nil
}

// GetDelistedStockByID returns a delisted stock by ID
func (r *DelistedStockRepo) GetDelistedStockByID(id int) (*models.DelistedStock, error) {
	query := `
		SELECT id, symbol, name, sector, final_price, delisted_at, reason
		FROM delisted_stocks
		WHERE id = ?
	`
	var stock models.DelistedStock
	var reasonStr string
	err := RetryQueryRow(r.db, query, id).Scan(
		&stock.ID,
		&stock.Symbol,
		&stock.Name,
		&stock.Sector,
		&stock.FinalPrice,
		&stock.DelistedAt,
		&reasonStr,
	)
	if err != nil {
		return nil, err
	}
	stock.Reason = models.DelistingReason(reasonStr)
	return &stock, nil
}

// GetRecentBankruptcies returns recent bankruptcies
func (r *DelistedStockRepo) GetRecentBankruptcies(limit int) ([]*models.DelistedStock, error) {
	query := `
		SELECT id, symbol, name, sector, final_price, delisted_at, reason
		FROM delisted_stocks
		WHERE reason = 'bankruptcy'
		ORDER BY delisted_at DESC
		LIMIT ?
	`
	rows, err := RetryQuery(r.db, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []*models.DelistedStock
	for rows.Next() {
		var stock models.DelistedStock
		var reasonStr string
		err := rows.Scan(
			&stock.ID,
			&stock.Symbol,
			&stock.Name,
			&stock.Sector,
			&stock.FinalPrice,
			&stock.DelistedAt,
			&reasonStr,
		)
		if err != nil {
			return nil, err
		}
		stock.Reason = models.DelistingReason(reasonStr)
		stocks = append(stocks, &stock)
	}
	return stocks, nil
}