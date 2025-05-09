package tests

import (
	"testing"

	"officestonks/internal/repository"
)

// TestUserRepositoryIntegration tests the user repository against a real database
func TestUserRepositoryIntegration(t *testing.T) {
	// Skip if no test database connection
	if TestDB == nil {
		t.Skip("No test database connection")
	}

	// Create repository
	userRepo := repository.NewUserRepo(TestDB)

	// Test username
	username := "integrationtestuser"
	password := "password123"

	// Test CreateUser
	t.Run("CreateUser", func(t *testing.T) {
		user, err := userRepo.CreateUser(username, password)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if user.ID <= 0 {
			t.Errorf("Expected positive user ID, got %d", user.ID)
		}
		if user.Username != username {
			t.Errorf("Expected username %s, got %s", username, user.Username)
		}
		if user.CashBalance != 10000.0 {
			t.Errorf("Expected initial cash balance 10000.0, got %f", user.CashBalance)
		}
	})

	// Test GetUserByUsername
	t.Run("GetUserByUsername", func(t *testing.T) {
		user, err := userRepo.GetUserByUsername(username)
		if err != nil {
			t.Fatalf("Failed to get user by username: %v", err)
		}

		if user.ID <= 0 {
			t.Errorf("Expected positive user ID, got %d", user.ID)
		}
		if user.Username != username {
			t.Errorf("Expected username %s, got %s", username, user.Username)
		}

		// Test non-existent username
		_, err = userRepo.GetUserByUsername("nonexistentuser")
		if err == nil {
			t.Error("Expected error when getting non-existent user, got nil")
		}
	})

	// Test UpdateUserBalance
	t.Run("UpdateUserBalance", func(t *testing.T) {
		// Get user by username first
		user, err := userRepo.GetUserByUsername(username)
		if err != nil {
			t.Fatalf("Failed to get user by username: %v", err)
		}

		// Update balance
		newBalance := 5000.0
		err = userRepo.UpdateUserBalance(user.ID, newBalance)
		if err != nil {
			t.Fatalf("Failed to update user balance: %v", err)
		}

		// Get user again to check balance
		updatedUser, err := userRepo.GetUserByUsername(username)
		if err != nil {
			t.Fatalf("Failed to get user by username after update: %v", err)
		}

		if updatedUser.CashBalance != newBalance {
			t.Errorf("Expected updated cash balance %f, got %f", newBalance, updatedUser.CashBalance)
		}
	})
}

// TestStockRepositoryIntegration tests the stock repository against a real database
func TestStockRepositoryIntegration(t *testing.T) {
	// Skip if no test database connection
	if TestDB == nil {
		t.Skip("No test database connection")
	}

	// Create repository
	stockRepo := repository.NewStockRepo(TestDB)

	// Test GetAllStocks
	t.Run("GetAllStocks", func(t *testing.T) {
		stocks, err := stockRepo.GetAllStocks()
		if err != nil {
			t.Fatalf("Failed to get all stocks: %v", err)
		}

		// Should have at least some stocks from the seed data
		if len(stocks) == 0 {
			t.Error("Expected some stocks, got none")
		}

		// Check stock fields
		for _, stock := range stocks {
			if stock.ID <= 0 {
				t.Errorf("Expected positive stock ID, got %d", stock.ID)
			}
			if stock.Symbol == "" {
				t.Error("Expected stock symbol, got empty string")
			}
			if stock.Name == "" {
				t.Error("Expected stock name, got empty string")
			}
			if stock.CurrentPrice <= 0 {
				t.Errorf("Expected positive stock price, got %f", stock.CurrentPrice)
			}
		}
	})

	// Other stock repository tests would go here:
	// - GetStockByID
	// - GetStockBySymbol
	// - UpdateStockPrice
	// - LoadStocksForSimulation
}

// Additional integration tests would be added for:
// - PortfolioRepository
// - TransactionRepository
// - MarketService (buying/selling stocks, etc.)
// - WebSocket functionality

// These tests are just a starting point and would need to be expanded
// for a complete test suite. They also depend on a properly initialized
// test database with seed data.