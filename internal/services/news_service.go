package services

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"officestonks/internal/models"
)

// NewsService handles business logic for news items and automated generation
type NewsService struct {
	repo models.NewsRepository
}

// NewNewsService creates a new NewsService
func NewNewsService(repo models.NewsRepository) *NewsService {
	return &NewsService{repo: repo}
}

// CreateNews adds a basic news item (for backward compatibility)
func (s *NewsService) CreateNews(title, content string, expiresAt time.Time) error {
	return s.repo.CreateNews(title, content, expiresAt)
}

// GetActiveNews retrieves non-expired news items
func (s *NewsService) GetActiveNews() ([]*models.NewsItem, error) {
	return s.repo.GetActiveNews()
}

// Crisis News Generation Methods

// GenerateCrisisNews creates news for a stock entering crisis at $0.01
func (s *NewsService) GenerateCrisisNews(stockID int, stockSymbol, stockName, sectorName string) error {
	templates := []string{
		"%s shares plummet to penny stock status amid investor concerns",
		"%s faces financial distress as stock hits rock bottom",
		"Market panic: %s trading at crisis levels",
		"%s stock collapses to $0.01 in heavy selling",
		"Investors flee %s as company enters crisis territory",
		"%s facing potential delisting as stock hits penny levels",
	}
	
	contentTemplates := []string{
		"Shares of %s (%s) have fallen to $0.01 amid growing concerns about the company's financial stability. The %s company now faces significant challenges as investors lose confidence.",
		"%s stock has crashed to penny stock levels, with shares trading at just $0.01. The dramatic decline has raised questions about the %s company's future prospects.",
		"In a dramatic turn of events, %s (%s) shares have collapsed to $0.01. The %s sector stock is now in crisis territory, with analysts warning of potential bankruptcy risks.",
		"Market turmoil continues for %s as the stock hits $0.01. The %s company's shares have lost nearly all their value, creating uncertainty for shareholders.",
	}
	
	title := fmt.Sprintf(templates[rand.Intn(len(templates))], stockName)
	content := fmt.Sprintf(contentTemplates[rand.Intn(len(contentTemplates))], stockName, stockSymbol, sectorName)
	
	newsItem := &models.NewsItem{
		Type:        models.NewsTypeCrisis,
		StockID:     &stockID,
		StockSymbol: &stockSymbol,
		SectorName:  &sectorName,
		Title:       title,
		Content:     content,
		ImpactType:  models.ImpactTypeImmediate,
		ImpactScore: -30, // Negative impact
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Expire in 24 hours
		IsAutomated: true,
	}
	
	log.Printf("📰 Generating crisis news for %s: %s", stockSymbol, title)
	return s.repo.CreateNewsItem(newsItem)
}

// GenerateBankruptcyNews creates news for a stock going bankrupt
func (s *NewsService) GenerateBankruptcyNews(stockID int, stockSymbol, stockName, sectorName string) error {
	templates := []string{
		"%s files for bankruptcy, stock to be delisted",
		"BREAKING: %s declares bankruptcy amid financial collapse",
		"%s bankruptcy filing shocks investors",
		"End of an era: %s files for Chapter 11 bankruptcy",
		"%s officially bankrupt, shareholders lose everything",
		"Bankruptcy confirmed: %s stock permanently delisted",
	}
	
	contentTemplates := []string{
		"%s (%s) has officially filed for bankruptcy protection, marking the end of the struggling %s company. All shareholders have lost their investments as the stock is permanently delisted from trading.",
		"In a devastating blow to investors, %s has filed for bankruptcy. The %s company's stock (%s) has been delisted, leaving shareholders with worthless shares.",
		"The worst fears have been realized as %s declares bankruptcy. The %s sector company's collapse represents a total loss for investors who held the stock through its crisis period.",
		"%s has become the latest casualty in the %s sector, filing for bankruptcy after failing to recover from its financial crisis. The stock (%s) is now worthless.",
	}
	
	title := fmt.Sprintf(templates[rand.Intn(len(templates))], stockName)
	content := fmt.Sprintf(contentTemplates[rand.Intn(len(contentTemplates))], stockName, stockSymbol, sectorName)
	
	newsItem := &models.NewsItem{
		Type:        models.NewsTypeBankruptcy,
		StockID:     &stockID,
		StockSymbol: &stockSymbol,
		SectorName:  &sectorName,
		Title:       title,
		Content:     content,
		ImpactType:  models.ImpactTypeImmediate,
		ImpactScore: -50, // Major negative impact
		ExpiresAt:   time.Now().Add(48 * time.Hour), // Expire in 48 hours
		IsAutomated: true,
	}
	
	log.Printf("📰 Generating bankruptcy news for %s: %s", stockSymbol, title)
	return s.repo.CreateNewsItem(newsItem)
}

// GenerateRecoveryNews creates news for a stock recovering from crisis
func (s *NewsService) GenerateRecoveryNews(stockID int, stockSymbol, stockName, sectorName string, newPrice float64) error {
	multiplier := newPrice / 0.01
	
	templates := []string{
		"MIRACLE: %s stages dramatic comeback with %.0fx surge",
		"Surprise acquisition saves %s from bankruptcy",
		"%s phoenix rises: Stock soars %.0fx in shocking recovery",
		"Against all odds: %s recovers with massive %.0fx jump",
		"Investor dreams come true: %s explodes %.0fx higher",
		"Stunning reversal: %s rockets %.0fx from penny stock status",
	}
	
	contentTemplates := []string{
		"In an extraordinary turn of events, %s (%s) has staged a dramatic recovery, surging %.0fx from its crisis low of $0.01 to $%.2f. The %s company's miraculous comeback has rewarded diamond-handed investors with massive returns.",
		"Dreams do come true on Wall Street. %s stock has exploded %.0fx higher to $%.2f after being rescued from bankruptcy. The %s company's recovery represents one of the most dramatic comebacks in recent market history.",
		"Penny stock millionaires are being minted as %s (%s) rockets %.0fx from $0.01 to $%.2f. The %s sector company's stunning reversal proves that even the most distressed stocks can reward patient investors.",
		"Breaking news: %s has achieved the impossible, rising %.0fx from crisis levels to $%.2f per share. The %s company's recovery story will go down in market legend, creating massive wealth for brave investors.",
	}
	
	title := fmt.Sprintf(templates[rand.Intn(len(templates))], stockName, multiplier)
	content := fmt.Sprintf(contentTemplates[rand.Intn(len(contentTemplates))], stockName, stockSymbol, multiplier, newPrice, sectorName)
	
	newsItem := &models.NewsItem{
		Type:        models.NewsTypeRecovery,
		StockID:     &stockID,
		StockSymbol: &stockSymbol,
		SectorName:  &sectorName,
		Title:       title,
		Content:     content,
		ImpactType:  models.ImpactTypeImmediate,
		ImpactScore: 40, // Strong positive impact
		ExpiresAt:   time.Now().Add(72 * time.Hour), // Expire in 72 hours (longer for good news)
		IsAutomated: true,
	}
	
	log.Printf("📰 Generating recovery news for %s: %s", stockSymbol, title)
	return s.repo.CreateNewsItem(newsItem)
}

// GenerateSectorContagionNews creates news for sector-wide impact
func (s *NewsService) GenerateSectorContagionNews(sectorName string, eventType string, affectedCount int) error {
	var templates []string
	var contentTemplates []string
	var impactScore int
	var newsType models.NewsType
	
	switch eventType {
	case "bankruptcy":
		templates = []string{
			"Crisis spreads: %s sector under pressure",
			"%s sector contagion as bankruptcy fears grow",
			"Panic selling hits %s stocks amid sector crisis",
			"%s sector tumbles on bankruptcy contagion",
		}
		contentTemplates = []string{
			"The bankruptcy crisis is spreading through the %s sector, with %d companies now showing signs of distress. Investors are fleeing the sector as contagion fears grow.",
			"Sector-wide panic has gripped %s stocks, with %d companies affected by the spreading crisis. The contagion effect is creating massive volatility across the sector.",
		}
		impactScore = -25
		newsType = models.NewsTypeSector
		
	case "recovery":
		templates = []string{
			"%s sector rallies on recovery optimism",
			"Positive sentiment lifts entire %s sector",
			"%s stocks surge on sector-wide recovery hopes",
			"Rising tide lifts all %s boats in sector rally",
		}
		contentTemplates = []string{
			"The miraculous recovery is spreading positive sentiment across the %s sector, with %d companies benefiting from renewed investor confidence.",
			"A sector-wide rally is underway in %s stocks, with %d companies seeing gains as investors embrace the recovery narrative.",
		}
		impactScore = 15
		newsType = models.NewsTypeSector
	}
	
	if len(templates) == 0 {
		return nil // Unknown event type
	}
	
	title := fmt.Sprintf(templates[rand.Intn(len(templates))], sectorName)
	content := fmt.Sprintf(contentTemplates[rand.Intn(len(contentTemplates))], sectorName, affectedCount)
	
	newsItem := &models.NewsItem{
		Type:        newsType,
		SectorName:  &sectorName,
		Title:       title,
		Content:     content,
		ImpactType:  models.ImpactTypeGradual,
		ImpactScore: impactScore,
		ExpiresAt:   time.Now().Add(36 * time.Hour), // Expire in 36 hours
		IsAutomated: true,
	}
	
	log.Printf("📰 Generating sector contagion news for %s: %s", sectorName, title)
	return s.repo.CreateNewsItem(newsItem)
}

// CleanupExpiredNews removes old news items
func (s *NewsService) CleanupExpiredNews() error {
	return s.repo.DeleteExpiredNews()
}

// Enhanced retrieval methods

// GetRecentNews returns recent news with limit
func (s *NewsService) GetRecentNews(limit int) ([]*models.NewsItem, error) {
	return s.repo.GetRecentNews(limit)
}

// GetCrisisNews returns recent crisis-related news
func (s *NewsService) GetCrisisNews(limit int) ([]*models.NewsItem, error) {
	return s.repo.GetNewsByType(models.NewsTypeCrisis, limit)
}

// GetStockNews returns news for a specific stock
func (s *NewsService) GetStockNews(stockID int, limit int) ([]*models.NewsItem, error) {
	return s.repo.GetNewsByStock(stockID, limit)
}

// GetSectorNews returns news for a specific sector
func (s *NewsService) GetSectorNews(sectorName string, limit int) ([]*models.NewsItem, error) {
	return s.repo.GetNewsBySector(sectorName, limit)
}