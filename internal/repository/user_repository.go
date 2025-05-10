package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"officestonks/internal/models"
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
		SELECT id, username, password_hash, cash_balance, is_admin, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CashBalance,
		&user.IsAdmin,
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
		SELECT id, username, password_hash, cash_balance, is_admin, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CashBalance,
		&user.IsAdmin,
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

// IsUserAdmin checks if a user is an admin
func (r *UserRepo) IsUserAdmin(userID int) (bool, error) {
	query := `
		SELECT is_admin
		FROM users
		WHERE id = ?
	`

	var isAdmin bool
	err := r.db.QueryRow(query, userID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("user not found")
		}
		return false, err
	}

	return isAdmin, nil
}

// GetAllUsers gets all users in the system
func (r *UserRepo) GetAllUsers() ([]*models.User, error) {
	query := `
		SELECT id, username, password_hash, cash_balance, is_admin, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`

	rows, err := r.db.Query(query)
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
			&user.PasswordHash,
			&user.CashBalance,
			&user.IsAdmin,
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

// UpdateUser updates a user's information
func (r *UserRepo) UpdateUser(userID int, cashBalance float64, isAdmin bool) error {
	query := `
		UPDATE users
		SET cash_balance = ?,
			is_admin = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, cashBalance, isAdmin, time.Now(), userID)
	return err
}

// DeleteUser deletes a user from the system
func (r *UserRepo) DeleteUser(userID int) error {
	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Delete user's transactions
	_, err = tx.Exec("DELETE FROM transactions WHERE user_id = ?", userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete user's portfolio
	_, err = tx.Exec("DELETE FROM portfolio WHERE user_id = ?", userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete user from chat
	_, err = tx.Exec("DELETE FROM chat_messages WHERE user_id = ?", userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Finally, delete the user
	_, err = tx.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
// For debugging purposes:
func (r *UserRepo) DebugIsUserAdmin(userID int) string {
	log.Printf("Debug IsUserAdmin: Checking admin status for user %d", userID)
	
	// Try to get the user first
	var user models.User
	userQuery := `
		SELECT id, username, password_hash, cash_balance, is_admin, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	err := r.db.QueryRow(userQuery, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CashBalance,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err \!= nil {
		if err == sql.ErrNoRows {
			return fmt.Sprintf("User %d not found", userID)
		}
		return fmt.Sprintf("Error getting user %d: %v", userID, err)
	}
	
	// Check the is_admin column directly
	var isAdmin bool
	adminQuery := `SELECT is_admin FROM users WHERE id = ?`
	err = r.db.QueryRow(adminQuery, userID).Scan(&isAdmin)
	
	if err \!= nil {
		return fmt.Sprintf("Error checking admin status for user %d: %v", userID, err)
	}
	
	// Check database schema
	var columnInfo string
	schemaQuery := `SHOW COLUMNS FROM users WHERE Field = 'is_admin'`
	err = r.db.QueryRow(schemaQuery).Scan(&columnInfo)
	
	if err \!= nil {
		columnInfo = fmt.Sprintf("Error checking is_admin column: %v", err)
	}
	
	return fmt.Sprintf("User %d: Username=%s, IsAdmin=%v (direct query: %v), Column info: %s", 
		userID, user.Username, user.IsAdmin, isAdmin, columnInfo)
}
EOF < /dev/null
