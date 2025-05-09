package repository

import (
	"database/sql"
	"time"

	"officestonks/internal/models"
)

// TransactionRepo implements the TransactionRepository interface
type TransactionRepo struct {
	db *sql.DB
}

// NewTransactionRepo creates a new transaction repository
func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

// CreateTransaction records a new transaction
func (r *TransactionRepo) CreateTransaction(userID, stockID, quantity int, price float64, transType models.TransactionType) (*models.Transaction, error) {
	query := `
		INSERT INTO transactions (user_id, stock_id, quantity, price, transaction_type)
		VALUES (?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, userID, stockID, quantity, price, transType)
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	transaction := &models.Transaction{
		ID:              int(id),
		UserID:          userID,
		StockID:         stockID,
		Quantity:        quantity,
		Price:           price,
		TransactionType: transType,
		CreatedAt:       time.Now(),
	}
	
	return transaction, nil
}

// GetUserTransactions gets a user's transaction history
func (r *TransactionRepo) GetUserTransactions(userID int, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT t.id, t.user_id, t.stock_id, t.quantity, t.price, t.transaction_type, t.created_at,
			   s.symbol, s.name
		FROM transactions t
		JOIN stocks s ON t.stock_id = s.id
		WHERE t.user_id = ?
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transactions []*models.Transaction
	for rows.Next() {
		var t models.Transaction
		var stock models.Stock
		
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.StockID,
			&t.Quantity,
			&t.Price,
			&t.TransactionType,
			&t.CreatedAt,
			&stock.Symbol,
			&stock.Name,
		)
		if err != nil {
			return nil, err
		}
		
		stock.ID = t.StockID
		t.Stock = stock
		transactions = append(transactions, &t)
	}
	
	return transactions, nil
}

// GetRecentTransactions gets the most recent transactions across all users
func (r *TransactionRepo) GetRecentTransactions(limit int) ([]*models.Transaction, error) {
	query := `
		SELECT t.id, t.user_id, t.stock_id, t.quantity, t.price, t.transaction_type, t.created_at,
			   s.symbol, s.name,
			   u.username
		FROM transactions t
		JOIN stocks s ON t.stock_id = s.id
		JOIN users u ON t.user_id = u.id
		ORDER BY t.created_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transactions []*models.Transaction
	for rows.Next() {
		var t models.Transaction
		var stock models.Stock
		var username string
		
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.StockID,
			&t.Quantity,
			&t.Price,
			&t.TransactionType,
			&t.CreatedAt,
			&stock.Symbol,
			&stock.Name,
			&username,
		)
		if err != nil {
			return nil, err
		}
		
		stock.ID = t.StockID
		t.Stock = stock
		transactions = append(transactions, &t)
	}
	
	return transactions, nil
}