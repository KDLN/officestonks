package repository

import (
	"database/sql"
	"time"

	"officestonks/internal/models"
)

// NewsRepo implements models.NewsRepository backed by MySQL
type NewsRepo struct {
	db *sql.DB
}

// NewNewsRepo creates a new NewsRepo
func NewNewsRepo(db *sql.DB) *NewsRepo {
	return &NewsRepo{db: db}
}

// CreateNews inserts a new news item
func (r *NewsRepo) CreateNews(title, content string, expiresAt time.Time) error {
	query := `
        INSERT INTO news (title, content, expires_at, created_at)
        VALUES (?, ?, ?, NOW())
    `
	_, err := r.db.Exec(query, title, content, expiresAt)
	return err
}

// GetActiveNews returns all news items that have not expired
func (r *NewsRepo) GetActiveNews() ([]*models.NewsItem, error) {
	query := `
        SELECT id, title, content, expires_at, created_at
        FROM news
        WHERE expires_at > NOW()
        ORDER BY created_at DESC
    `
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.NewsItem
	for rows.Next() {
		var n models.NewsItem
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.ExpiresAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &n)
	}
	return items, nil
}
