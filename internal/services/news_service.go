package services

import (
	"time"

	"officestonks/internal/models"
)

// NewsService handles business logic for news items
type NewsService struct {
	repo models.NewsRepository
}

// NewNewsService creates a new NewsService
func NewNewsService(repo models.NewsRepository) *NewsService {
	return &NewsService{repo: repo}
}

// CreateNews adds a new news item
func (s *NewsService) CreateNews(title, content string, expiresAt time.Time) error {
	return s.repo.CreateNews(title, content, expiresAt)
}

// GetActiveNews retrieves non-expired news items
func (s *NewsService) GetActiveNews() ([]*models.NewsItem, error) {
	return s.repo.GetActiveNews()
}
