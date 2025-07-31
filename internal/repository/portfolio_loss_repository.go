package repository

import (
	"database/sql"

	"officestonks/internal/models"
)

// PortfolioLossRepo implements models.PortfolioLossRepository backed by MySQL
type PortfolioLossRepo struct {
	db *sql.DB
}

// NewPortfolioLossRepo creates a new PortfolioLossRepo
func NewPortfolioLossRepo(db *sql.DB) *PortfolioLossRepo {
	return &PortfolioLossRepo{db: db}
}

// CreatePortfolioLoss records a portfolio loss from bankruptcy
func (r *PortfolioLossRepo) CreatePortfolioLoss(userID, stockID int, stockSymbol, stockName string, quantity int, lossAmount float64) error {
	query := `
		INSERT INTO portfolio_losses (user_id, stock_id, stock_symbol, stock_name, quantity, loss_amount, delisted_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := RetryExec(r.db, query, userID, stockID, stockSymbol, stockName, quantity, lossAmount)
	return err
}

// GetUserLosses returns paginated losses for a user
func (r *PortfolioLossRepo) GetUserLosses(userID int, limit, offset int) ([]*models.PortfolioLoss, error) {
	query := `
		SELECT id, user_id, stock_id, stock_symbol, stock_name, quantity, loss_amount, delisted_at
		FROM portfolio_losses
		WHERE user_id = ?
		ORDER BY delisted_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := RetryQuery(r.db, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var losses []*models.PortfolioLoss
	for rows.Next() {
		var loss models.PortfolioLoss
		err := rows.Scan(
			&loss.ID,
			&loss.UserID,
			&loss.StockID,
			&loss.StockSymbol,
			&loss.StockName,
			&loss.Quantity,
			&loss.LossAmount,
			&loss.DelistedAt,
		)
		if err != nil {
			return nil, err
		}
		losses = append(losses, &loss)
	}
	return losses, nil
}

// GetTotalUserLosses returns the total amount a user has lost to bankruptcies
func (r *PortfolioLossRepo) GetTotalUserLosses(userID int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(loss_amount), 0)
		FROM portfolio_losses
		WHERE user_id = ?
	`
	var total float64
	err := RetryQueryRow(r.db, query, userID).Scan(&total)
	return total, err
}

// GetRecentLosses returns recent losses across all users
func (r *PortfolioLossRepo) GetRecentLosses(limit int) ([]*models.PortfolioLoss, error) {
	query := `
		SELECT pl.id, pl.user_id, pl.stock_id, pl.stock_symbol, pl.stock_name, 
		       pl.quantity, pl.loss_amount, pl.delisted_at
		FROM portfolio_losses pl
		ORDER BY pl.delisted_at DESC
		LIMIT ?
	`
	rows, err := RetryQuery(r.db, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var losses []*models.PortfolioLoss
	for rows.Next() {
		var loss models.PortfolioLoss
		err := rows.Scan(
			&loss.ID,
			&loss.UserID,
			&loss.StockID,
			&loss.StockSymbol,
			&loss.StockName,
			&loss.Quantity,
			&loss.LossAmount,
			&loss.DelistedAt,
		)
		if err != nil {
			return nil, err
		}
		losses = append(losses, &loss)
	}
	return losses, nil
}

// GetLossesByStock returns losses for a specific stock
func (r *PortfolioLossRepo) GetLossesByStock(stockID int, limit int) ([]*models.PortfolioLoss, error) {
	query := `
		SELECT id, user_id, stock_id, stock_symbol, stock_name, quantity, loss_amount, delisted_at
		FROM portfolio_losses
		WHERE stock_id = ?
		ORDER BY delisted_at DESC
		LIMIT ?
	`
	rows, err := RetryQuery(r.db, query, stockID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var losses []*models.PortfolioLoss
	for rows.Next() {
		var loss models.PortfolioLoss
		err := rows.Scan(
			&loss.ID,
			&loss.UserID,
			&loss.StockID,
			&loss.StockSymbol,
			&loss.StockName,
			&loss.Quantity,
			&loss.LossAmount,
			&loss.DelistedAt,
		)
		if err != nil {
			return nil, err
		}
		losses = append(losses, &loss)
	}
	return losses, nil
}