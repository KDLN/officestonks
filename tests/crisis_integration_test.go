package tests

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"officestonks/internal/models"
	"officestonks/internal/repository"
	"officestonks/internal/services"
)

// TestCrisisMechanics tests the complete crisis/bankruptcy/recovery flow
func TestCrisisMechanics(t *testing.T) {
	// Setup test database connection
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Create repositories
	stockRepo := repository.NewStockRepo(db)
	userRepo := repository.NewUserRepo(db)
	portfolioRepo := repository.NewPortfolioRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)
	sectorRepo := repository.NewSectorRepo(db)
	delistedStockRepo := repository.NewDelistedStockRepo(db)
	portfolioLossRepo := repository.NewPortfolioLossRepo(db)
	newsRepo := repository.NewNewsRepo(db)

	// Create services
	newsService := services.NewNewsService(newsRepo)
	marketService := services.NewMarketService(
		stockRepo, userRepo, portfolioRepo, transactionRepo, sectorRepo,
		delistedStockRepo, portfolioLossRepo, newsService,
	)

	// Initialize the simulator
	err = marketService.InitializeSimulator()
	if err != nil {
		t.Fatalf("Failed to initialize simulator: %v", err)
	}

	// Test 1: Create test users with portfolios
	t.Run("Setup Test Users", func(t *testing.T) {
		// Create test users
		users := []struct {
			username string
			cash     float64
		}{
			{"test_user_1", 10000},
			{"test_user_2", 15000},
			{"test_user_3", 20000},
		}

		for _, u := range users {
			_, err := userRepo.CreateUser(u.username, "password123")
			if err != nil {
				t.Errorf("Failed to create user %s: %v", u.username, err)
			}
		}
	})

	// Test 2: Add stocks to user portfolios
	t.Run("Add Stocks to Portfolios", func(t *testing.T) {
		// Get test users
		user1, _ := userRepo.GetUserByUsername("test_user_1")
		user2, _ := userRepo.GetUserByUsername("test_user_2")

		// Add stocks to portfolios
		portfolioRepo.AddStockToPortfolio(user1.ID, 1, 100) // 100 shares of stock 1
		portfolioRepo.AddStockToPortfolio(user1.ID, 2, 50)  // 50 shares of stock 2
		portfolioRepo.AddStockToPortfolio(user2.ID, 1, 200) // 200 shares of stock 1
		portfolioRepo.AddStockToPortfolio(user2.ID, 3, 150) // 150 shares of stock 3
	})

	// Test 3: Force Crisis Event
	t.Run("Force Crisis Event", func(t *testing.T) {
		err := marketService.ForceCrisisEvent(1)
		if err != nil {
			t.Errorf("Failed to force crisis event: %v", err)
		}

		// Wait for processing
		time.Sleep(2 * time.Second)

		// Verify stock is in crisis
		stock, _ := stockRepo.GetStockByID(1)
		if stock.CurrentPrice != 0.01 {
			t.Errorf("Stock not at crisis price. Expected: 0.01, Got: %f", stock.CurrentPrice)
		}
	})

	// Test 4: Force Bankruptcy with Portfolio Impact
	t.Run("Force Bankruptcy", func(t *testing.T) {
		// Get holdings before bankruptcy
		holdingsBefore, _ := portfolioRepo.GetAllHoldersOfStock(1)
		t.Logf("Holdings before bankruptcy: %d users", len(holdingsBefore))

		// Force bankruptcy
		err := marketService.ForceBankruptcy(1)
		if err != nil {
			t.Errorf("Failed to force bankruptcy: %v", err)
		}

		// Wait for processing
		time.Sleep(2 * time.Second)

		// Verify bankruptcy processing
		// 1. Check holdings removed
		holdingsAfter, _ := portfolioRepo.GetAllHoldersOfStock(1)
		if len(holdingsAfter) != 0 {
			t.Errorf("Holdings not removed. Expected: 0, Got: %d", len(holdingsAfter))
		}

		// 2. Check portfolio losses recorded
		user1, _ := userRepo.GetUserByUsername("test_user_1")
		losses, _ := portfolioLossRepo.GetUserLosses(user1.ID, 10, 0)
		if len(losses) == 0 {
			t.Error("No portfolio losses recorded for user")
		}

		// 3. Check stock delisted
		delistedStocks, _ := delistedStockRepo.GetRecentBankruptcies(10)
		found := false
		for _, ds := range delistedStocks {
			if ds.ID == 1 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Stock not found in delisted_stocks table")
		}

		// 4. Check news generated
		news, _ := newsRepo.GetLatestNews(10, 0)
		bankruptcyNewsFound := false
		for _, n := range news {
			if n.Type == models.NewsTypeBankruptcy && n.StockID != nil && *n.StockID == 1 {
				bankruptcyNewsFound = true
				break
			}
		}
		if !bankruptcyNewsFound {
			t.Error("Bankruptcy news not generated")
		}
	})

	// Test 5: Sector Contagion
	t.Run("Sector Contagion", func(t *testing.T) {
		// Get stocks in same sector as bankrupt stock
		stocks, _ := stockRepo.GetAllStocks()
		var sectorStocks []int
		bankruptStock, _ := stockRepo.GetStockByID(1)

		for _, s := range stocks {
			if s.Sector == bankruptStock.Sector && s.ID != 1 {
				sectorStocks = append(sectorStocks, s.ID)
			}
		}

		// Check if any sector stocks are affected
		affectedCount := 0
		for _, stockID := range sectorStocks {
			stock, _ := stockRepo.GetStockByID(stockID)
			if stock.Status == models.StockDistressed || stock.CurrentPrice < 10.0 {
				affectedCount++
			}
		}

		t.Logf("Sector contagion: %d/%d stocks affected", affectedCount, len(sectorStocks))
	})

	// Test 6: Force Recovery
	t.Run("Force Recovery", func(t *testing.T) {
		// First, put a stock in crisis
		marketService.ForceCrisisEvent(3)
		time.Sleep(1 * time.Second)

		// Force recovery
		err := marketService.ForceRecovery(3)
		if err != nil {
			t.Errorf("Failed to force recovery: %v", err)
		}

		// Wait for processing
		time.Sleep(2 * time.Second)

		// Verify recovery
		stock, _ := stockRepo.GetStockByID(3)
		if stock.CurrentPrice <= 0.01 {
			t.Errorf("Stock did not recover. Price: %f", stock.CurrentPrice)
		}

		// Check recovery news
		news, _ := newsRepo.GetLatestNews(10, 0)
		recoveryNewsFound := false
		for _, n := range news {
			if n.Type == models.NewsTypeRecovery && n.StockID != nil && *n.StockID == 3 {
				recoveryNewsFound = true
				break
			}
		}
		if !recoveryNewsFound {
			t.Error("Recovery news not generated")
		}
	})

	// Cleanup
	marketService.PauseSimulator()
}

// Helper function to setup test database
func setupTestDB() (*sql.DB, error) {
	// This would connect to a test database
	// For now, return error to indicate this needs implementation
	return nil, fmt.Errorf("test database setup not implemented")
}

// TestSectorCascade tests cascading failures in a sector
func TestSectorCascade(t *testing.T) {
	// This would test multiple bankruptcies triggering sector-wide crisis
	t.Skip("Implement sector cascade testing")
}

// TestPortfolioLossCalculation verifies loss calculations are correct
func TestPortfolioLossCalculation(t *testing.T) {
	// This would verify that loss amounts are calculated correctly
	t.Skip("Implement loss calculation testing")
}