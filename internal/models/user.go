package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose this in JSON
	CashBalance  float64   `json:"cash_balance"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository interface defines methods for user data access
type UserRepository interface {
	CreateUser(username, password string) (*User, error)
	GetUserByID(id int) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUserBalance(userID int, newBalance float64) error
	GetTopUsers(limit int) ([]*User, error)
	IsUserAdmin(userID int) (bool, error)
	GetAllUsers() ([]*User, error)
	UpdateUser(userID int, cashBalance float64, isAdmin bool) error
	DeleteUser(userID int) error
}

// AuthRequest is used for login/register requests
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse is sent after successful authentication
type AuthResponse struct {
	Token   string `json:"token"`
	UserID  int    `json:"user_id"`
	IsAdmin bool   `json:"is_admin"`
}