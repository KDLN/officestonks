package models

import (
	"time"
)

// ChatMessage represents a chat message in the system
type ChatMessage struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatRepository interface defines methods for chat data access
type ChatRepository interface {
	SaveMessage(userID int, message string) (*ChatMessage, error)
	GetRecentMessages(limit int) ([]*ChatMessage, error)
	ClearAllMessages() error
}