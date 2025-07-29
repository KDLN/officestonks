package models

import "time"

// NewsItem represents a news post shared by admins
type NewsItem struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewsRepository defines data access methods for news
type NewsRepository interface {
	CreateNews(title, content string, expiresAt time.Time) error
	GetActiveNews() ([]*NewsItem, error)
}
