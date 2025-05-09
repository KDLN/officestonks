package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/yourusername/officestonks/internal/models"
)

// UserRepo implements the UserRepository interface
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new user repository
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// CreateUser adds a new user to the database
func (r *UserRepo) CreateUser(username, passwordHash string) (*models.User, error) {
	// Initial cash balance
	initialBalance := 10000.00
	
	// SQL statement to insert a new user
	query := `
		INSERT INTO users (username, password_hash, cash_balance)
		VALUES (?, ?, ?)
	`
	
	// Execute the query
	result, err := r.db.Exec(query, username, passwordHash, initialBalance)
	if err != nil {
		return nil, err
	}
	
	// Get the ID of the new user
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// Return the new user
	return &models.User{
		ID:           int(id),
		Username:     username,
		PasswordHash: passwordHash,
		CashBalance:  initialBalance,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepo) GetUserByID(id int) (*models.User, error) {
	var user models.User
	
	query := `
		SELECT id, username, password_hash, cash_balance, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CashBalance,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepo) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	
	query := `
		SELECT id, username, password_hash, cash_balance, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CashBalance,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	
	return &user, nil
}

// UpdateUserBalance updates a user's cash balance
func (r *UserRepo) UpdateUserBalance(userID int, newBalance float64) error {
	query := `
		UPDATE users
		SET cash_balance = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, newBalance, userID)
	return err
}

// GetTopUsers gets the top users by portfolio value
func (r *UserRepo) GetTopUsers(limit int) ([]*models.User, error) {
	// This is simplified - in a real app, you'd calculate portfolio value
	query := `
		SELECT id, username, cash_balance, created_at, updated_at
		FROM users
		ORDER BY cash_balance DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.CashBalance,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	
	return users, nil
}