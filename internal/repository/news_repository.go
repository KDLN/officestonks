package repository

import (
	"database/sql"
	"fmt"
	"log"
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

// CreateNews inserts a basic news item (for backward compatibility)
func (r *NewsRepo) CreateNews(title, content string, expiresAt time.Time) error {
	newsItem := &models.NewsItem{
		Type:        models.NewsTypeMarket,
		Title:       title,
		Content:     content,
		ImpactType:  models.ImpactTypeGradual,
		ImpactScore: 0,
		ExpiresAt:   expiresAt,
		IsAutomated: false,
	}
	return r.CreateNewsItem(newsItem)
}

// CreateNewsItem inserts an enhanced news item with all fields
func (r *NewsRepo) CreateNewsItem(newsItem *models.NewsItem) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
        INSERT INTO news_items (type, stock_id, stock_symbol, sector_name, title, content, 
                              impact_type, impact_score, expires_at, is_automated)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	
	_, err := r.db.Exec(query, 
		string(newsItem.Type),
		newsItem.StockID,
		newsItem.StockSymbol,
		newsItem.SectorName,
		newsItem.Title,
		newsItem.Content,
		string(newsItem.ImpactType),
		newsItem.ImpactScore,
		newsItem.ExpiresAt,
		newsItem.IsAutomated,
	)
	
	if err != nil {
		log.Printf("📰 Error creating news item: %v", err)
		return err
	}
	
	log.Printf("📰 Created news: %s [%s]", newsItem.Title, string(newsItem.Type))
	return nil
}

// GetActiveNews returns all active news items (for backward compatibility)
func (r *NewsRepo) GetActiveNews() ([]*models.NewsItem, error) {
	return r.GetRecentNews(50)
}

// GetRecentNews returns recent news items with limit
func (r *NewsRepo) GetRecentNews(limit int) ([]*models.NewsItem, error) {
	if r.db == nil {
		return []*models.NewsItem{}, nil
	}

	query := `
        SELECT id, type, stock_id, stock_symbol, sector_name, title, content, 
               impact_type, impact_score, created_at, expires_at, is_automated
        FROM news_items
        WHERE expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT ?
    `
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		log.Printf("📰 Error querying recent news: %v", err)
		return []*models.NewsItem{}, nil
	}
	defer rows.Close()

	return r.scanNewsItems(rows)
}

// GetNewsByType returns news items of a specific type
func (r *NewsRepo) GetNewsByType(newsType models.NewsType, limit int) ([]*models.NewsItem, error) {
	if r.db == nil {
		return []*models.NewsItem{}, nil
	}

	query := `
        SELECT id, type, stock_id, stock_symbol, sector_name, title, content, 
               impact_type, impact_score, created_at, expires_at, is_automated
        FROM news_items
        WHERE type = ? AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT ?
    `
	
	rows, err := r.db.Query(query, string(newsType), limit)
	if err != nil {
		log.Printf("📰 Error querying news by type %s: %v", newsType, err)
		return []*models.NewsItem{}, nil
	}
	defer rows.Close()

	return r.scanNewsItems(rows)
}

// GetNewsByStock returns news items for a specific stock
func (r *NewsRepo) GetNewsByStock(stockID int, limit int) ([]*models.NewsItem, error) {
	if r.db == nil {
		return []*models.NewsItem{}, nil
	}

	query := `
        SELECT id, type, stock_id, stock_symbol, sector_name, title, content, 
               impact_type, impact_score, created_at, expires_at, is_automated
        FROM news_items
        WHERE stock_id = ? AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT ?
    `
	
	rows, err := r.db.Query(query, stockID, limit)
	if err != nil {
		log.Printf("📰 Error querying news for stock %d: %v", stockID, err)
		return []*models.NewsItem{}, nil
	}
	defer rows.Close()

	return r.scanNewsItems(rows)
}

// GetNewsBySector returns news items for a specific sector
func (r *NewsRepo) GetNewsBySector(sectorName string, limit int) ([]*models.NewsItem, error) {
	if r.db == nil {
		return []*models.NewsItem{}, nil
	}

	query := `
        SELECT id, type, stock_id, stock_symbol, sector_name, title, content, 
               impact_type, impact_score, created_at, expires_at, is_automated
        FROM news_items
        WHERE sector_name = ? AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT ?
    `
	
	rows, err := r.db.Query(query, sectorName, limit)
	if err != nil {
		log.Printf("📰 Error querying news for sector %s: %v", sectorName, err)
		return []*models.NewsItem{}, nil
	}
	defer rows.Close()

	return r.scanNewsItems(rows)
}

// DeleteExpiredNews removes expired news items
func (r *NewsRepo) DeleteExpiredNews() error {
	if r.db == nil {
		return nil
	}

	query := `DELETE FROM news_items WHERE expires_at <= NOW()`
	
	result, err := r.db.Exec(query)
	if err != nil {
		log.Printf("📰 Error deleting expired news: %v", err)
		return err
	}
	
	count, _ := result.RowsAffected()
	if count > 0 {
		log.Printf("📰 Deleted %d expired news items", count)
	}
	
	return nil
}

// scanNewsItems is a helper method to scan rows into NewsItem structs
func (r *NewsRepo) scanNewsItems(rows *sql.Rows) ([]*models.NewsItem, error) {
	var items []*models.NewsItem
	
	for rows.Next() {
		var n models.NewsItem
		var stockSymbol, sectorName sql.NullString
		var stockIDInt sql.NullInt64
		
		err := rows.Scan(
			&n.ID,
			&n.Type,
			&stockIDInt,
			&stockSymbol,
			&sectorName,
			&n.Title,
			&n.Content,
			&n.ImpactType,
			&n.ImpactScore,
			&n.CreatedAt,
			&n.ExpiresAt,
			&n.IsAutomated,
		)
		
		if err != nil {
			log.Printf("📰 Error scanning news item: %v", err)
			continue
		}
		
		// Handle nullable fields
		if stockIDInt.Valid {
			stockIDValue := int(stockIDInt.Int64)
			n.StockID = &stockIDValue
		}
		if stockSymbol.Valid {
			n.StockSymbol = &stockSymbol.String
		}
		if sectorName.Valid {
			n.SectorName = &sectorName.String
		}
		
		items = append(items, &n)
	}
	
	return items, nil
}