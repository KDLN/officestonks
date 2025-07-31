package models

import "time"

// NewsType represents the type of news item
type NewsType string

const (
	NewsTypeCompany    NewsType = "company"
	NewsTypeSector     NewsType = "sector"
	NewsTypeMarket     NewsType = "market"
	NewsTypeCrisis     NewsType = "crisis"
	NewsTypeRecovery   NewsType = "recovery"
	NewsTypeBankruptcy NewsType = "bankruptcy"
)

// ImpactType represents how the news affects market prices
type ImpactType string

const (
	ImpactTypeImmediate ImpactType = "immediate"
	ImpactTypeGradual   ImpactType = "gradual"
)

// NewsItem represents a news item with market impact
type NewsItem struct {
	ID          int        `json:"id"`
	Type        NewsType   `json:"type"`
	StockID     *int       `json:"stock_id,omitempty"`
	StockSymbol *string    `json:"stock_symbol,omitempty"`
	SectorName  *string    `json:"sector_name,omitempty"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ImpactType  ImpactType `json:"impact_type"`
	ImpactScore int        `json:"impact_score"` // -100 to +100
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	IsAutomated bool       `json:"is_automated"`
}

// NewsRepository defines data access methods for news
type NewsRepository interface {
	// Basic news operations
	CreateNews(title, content string, expiresAt time.Time) error
	GetActiveNews() ([]*NewsItem, error)
	
	// Enhanced news operations for crisis system
	CreateNewsItem(newsItem *NewsItem) error
	GetRecentNews(limit int) ([]*NewsItem, error)
	GetNewsByType(newsType NewsType, limit int) ([]*NewsItem, error)
	GetNewsByStock(stockID int, limit int) ([]*NewsItem, error)
	GetNewsBySector(sectorName string, limit int) ([]*NewsItem, error)
	DeleteExpiredNews() error
}
