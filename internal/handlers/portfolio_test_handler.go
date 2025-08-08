package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"officestonks/internal/models"
	"officestonks/internal/services"
)

// PortfolioTestHandler handles portfolio and trading test scenarios
type PortfolioTestHandler struct {
	marketService     *services.MarketService
	userService       *services.UserService
	stockRepo         models.StockRepository
	userRepo          models.UserRepository
	portfolioRepo     models.PortfolioRepository
	transactionRepo   models.TransactionRepository
}

// NewPortfolioTestHandler creates a new portfolio test handler
func NewPortfolioTestHandler(
	marketService *services.MarketService,
	userService *services.UserService,
	stockRepo models.StockRepository,
	userRepo models.UserRepository,
	portfolioRepo models.PortfolioRepository,
	transactionRepo models.TransactionRepository,
) *PortfolioTestHandler {
	return &PortfolioTestHandler{
		marketService:   marketService,
		userService:     userService,
		stockRepo:       stockRepo,
		userRepo:        userRepo,
		portfolioRepo:   portfolioRepo,
		transactionRepo: transactionRepo,
	}
}

// RunPortfolioTests executes the portfolio and trading test suite
func (h *PortfolioTestHandler) RunPortfolioTests(w http.ResponseWriter, r *http.Request) {
	suite := TestSuite{
		SuiteName: "Portfolio & Trading Test Suite",
		StartTime: time.Now(),
		Tests:     []TestResult{},
	}

	// Test 1: Portfolio Calculation Accuracy
	h.runPortfolioTest(&suite, "Portfolio Calculation Accuracy", func() (map[string]interface{}, error) {
		// Create test user
		testUser, err := h.createTestUser("portfolio_calc_user", 10000)
		if err != nil {
			return nil, err
		}
		defer h.cleanupTestUser(testUser.ID)

		// Get a healthy stock
		stock, err := h.getHealthyStock()
		if err != nil {
			return nil, err
		}

		// Buy shares
		quantity := 10
		err = h.marketService.BuyStock(testUser.ID, stock.ID, quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to buy stock: %w", err)
		}

		// Get portfolio
		portfolio, err := h.marketService.GetUserPortfolio(testUser.ID)
		if err != nil {
			return nil, err
		}

		// Calculate expected values
		expectedStockValue := float64(quantity) * stock.CurrentPrice
		expectedTotalValue := (10000 - expectedStockValue) + expectedStockValue // Cash + Stock

		// Verify calculations
		stockValueMatch := abs(portfolio.StockValue-expectedStockValue) < 0.01
		totalValueMatch := abs(portfolio.TotalValue-expectedTotalValue) < 0.01

		return map[string]interface{}{
			"user_id":                testUser.ID,
			"stock_purchased":        stock.Symbol,
			"quantity":               quantity,
			"stock_price":            stock.CurrentPrice,
			"expected_stock_value":   expectedStockValue,
			"actual_stock_value":     portfolio.StockValue,
			"stock_value_accurate":   stockValueMatch,
			"expected_total_value":   expectedTotalValue,
			"actual_total_value":     portfolio.TotalValue,
			"total_value_accurate":   totalValueMatch,
			"cash_balance":           portfolio.CashBalance,
			"portfolio_items_count":  len(portfolio.PortfolioItems),
		}, nil
	})

	// Test 2: Buy Order Processing
	h.runPortfolioTest(&suite, "Buy Order Processing", func() (map[string]interface{}, error) {
		testUser, err := h.createTestUser("buy_order_user", 5000)
		if err != nil {
			return nil, err
		}
		defer h.cleanupTestUser(testUser.ID)

		stock, err := h.getHealthyStock()
		if err != nil {
			return nil, err
		}

		initialCash := testUser.CashBalance
		quantity := 5
		expectedCost := float64(quantity) * stock.CurrentPrice

		// Execute buy order
		err = h.marketService.BuyStock(testUser.ID, stock.ID, quantity)
		if err != nil {
			return nil, err
		}

		// Verify user cash updated
		updatedUser, _ := h.userRepo.GetUserByID(testUser.ID)
		expectedCash := initialCash - expectedCost
		cashCorrect := abs(updatedUser.CashBalance-expectedCash) < 0.01

		// Verify portfolio holding created
		holding, _ := h.portfolioRepo.GetUserStockHolding(testUser.ID, stock.ID)
		holdingCorrect := holding != nil && holding.Quantity == quantity

		// Verify transaction recorded
		transactions, _ := h.transactionRepo.GetUserTransactions(testUser.ID, 10, 0)
		transactionFound := false
		for _, tx := range transactions {
			if tx.StockID == stock.ID && tx.Quantity == quantity && tx.TransactionType == models.Buy {
				transactionFound = true
				break
			}
		}

		return map[string]interface{}{
			"stock_symbol":        stock.Symbol,
			"quantity_bought":     quantity,
			"cost_per_share":      stock.CurrentPrice,
			"total_cost":          expectedCost,
			"initial_cash":        initialCash,
			"final_cash":          updatedUser.CashBalance,
			"cash_updated_correctly": cashCorrect,
			"holding_created":     holdingCorrect,
			"transaction_recorded": transactionFound,
		}, nil
	})

	// Test 3: Sell Order Processing
	h.runPortfolioTest(&suite, "Sell Order Processing", func() (map[string]interface{}, error) {
		testUser, err := h.createTestUser("sell_order_user", 8000)
		if err != nil {
			return nil, err
		}
		defer h.cleanupTestUser(testUser.ID)

		stock, err := h.getHealthyStock()
		if err != nil {
			return nil, err
		}

		// First buy some shares
		buyQuantity := 10
		err = h.marketService.BuyStock(testUser.ID, stock.ID, buyQuantity)
		if err != nil {
			return nil, err
		}

		// Get cash after buy
		userAfterBuy, _ := h.userRepo.GetUserByID(testUser.ID)
		cashAfterBuy := userAfterBuy.CashBalance

		// Now sell some shares
		sellQuantity := 6
		err = h.marketService.SellStock(testUser.ID, stock.ID, sellQuantity)
		if err != nil {
			return nil, err
		}

		// Verify results
		finalUser, _ := h.userRepo.GetUserByID(testUser.ID)
		expectedProceeds := float64(sellQuantity) * stock.CurrentPrice
		expectedFinalCash := cashAfterBuy + expectedProceeds
		cashCorrect := abs(finalUser.CashBalance-expectedFinalCash) < 0.01

		// Check remaining holding
		holding, _ := h.portfolioRepo.GetUserStockHolding(testUser.ID, stock.ID)
		remainingShares := buyQuantity - sellQuantity
		holdingCorrect := holding != nil && holding.Quantity == remainingShares

		// Check sell transaction
		transactions, _ := h.transactionRepo.GetUserTransactions(testUser.ID, 10, 0)
		sellTransactionFound := false
		for _, tx := range transactions {
			if tx.StockID == stock.ID && tx.Quantity == sellQuantity && tx.TransactionType == models.Sell {
				sellTransactionFound = true
				break
			}
		}

		return map[string]interface{}{
			"stock_symbol":           stock.Symbol,
			"initial_quantity":       buyQuantity,
			"sold_quantity":          sellQuantity,
			"remaining_quantity":     remainingShares,
			"sell_price_per_share":   stock.CurrentPrice,
			"expected_proceeds":      expectedProceeds,
			"cash_after_buy":         cashAfterBuy,
			"final_cash":             finalUser.CashBalance,
			"cash_updated_correctly": cashCorrect,
			"holding_updated":        holdingCorrect,
			"sell_transaction_recorded": sellTransactionFound,
		}, nil
	})

	// Test 4: Insufficient Funds Handling
	h.runPortfolioTest(&suite, "Insufficient Funds Handling", func() (map[string]interface{}, error) {
		testUser, err := h.createTestUser("poor_user", 100) // Very low cash
		if err != nil {
			return nil, err
		}
		defer h.cleanupTestUser(testUser.ID)

		stock, err := h.getHealthyStock()
		if err != nil {
			return nil, err
		}

		// Try to buy more than user can afford
		expensiveQuantity := int(200 / stock.CurrentPrice) + 1 // More than $100 allows
		
		err = h.marketService.BuyStock(testUser.ID, stock.ID, expensiveQuantity)
		insufficientFundsError := err != nil && err.Error() == "insufficient funds"

		// Verify no changes occurred
		finalUser, _ := h.userRepo.GetUserByID(testUser.ID)
		cashUnchanged := finalUser.CashBalance == 100

		holding, _ := h.portfolioRepo.GetUserStockHolding(testUser.ID, stock.ID)
		noHoldingCreated := holding == nil

		return map[string]interface{}{
			"user_cash":                 100.0,
			"stock_price":              stock.CurrentPrice,
			"attempted_quantity":       expensiveQuantity,
			"attempted_cost":           float64(expensiveQuantity) * stock.CurrentPrice,
			"insufficient_funds_error": insufficientFundsError,
			"cash_unchanged":           cashUnchanged,
			"no_holding_created":       noHoldingCreated,
		}, nil
	})

	// Test 5: Share Quantity Validation
	h.runPortfolioTest(&suite, "Share Quantity Validation", func() (map[string]interface{}, error) {
		testUser, err := h.createTestUser("share_validation_user", 5000)
		if err != nil {
			return nil, err
		}
		defer h.cleanupTestUser(testUser.ID)

		stock, err := h.getHealthyStock()
		if err != nil {
			return nil, err
		}

		// Test invalid quantities
		tests := map[string]struct {
			quantity int
			action   string
		}{
			"zero_buy":     {0, "buy"},
			"negative_buy": {-5, "buy"},
			"zero_sell":    {0, "sell"},
			"negative_sell": {-3, "sell"},
		}

		results := make(map[string]bool)
		
		for testName, test := range tests {
			var err error
			if test.action == "buy" {
				err = h.marketService.BuyStock(testUser.ID, stock.ID, test.quantity)
			} else {
				err = h.marketService.SellStock(testUser.ID, stock.ID, test.quantity)
			}
			results[testName+"_rejected"] = err != nil
		}

		return map[string]interface{}{
			"validation_tests": results,
		}, nil
	})

	// Test 6: Transaction History Integrity
	h.runPortfolioTest(&suite, "Transaction History Integrity", func() (map[string]interface{}, error) {
		testUser, err := h.createTestUser("transaction_user", 10000)
		if err != nil {
			return nil, err
		}
		defer h.cleanupTestUser(testUser.ID)

		stock, err := h.getHealthyStock()
		if err != nil {
			return nil, err
		}

		// Perform multiple transactions
		transactions := []struct {
			action   string
			quantity int
		}{
			{"buy", 5},
			{"buy", 3},
			{"sell", 2},
			{"buy", 1},
			{"sell", 4},
		}

		for _, tx := range transactions {
			if tx.action == "buy" {
				h.marketService.BuyStock(testUser.ID, stock.ID, tx.quantity)
			} else {
				h.marketService.SellStock(testUser.ID, stock.ID, tx.quantity)
			}
			time.Sleep(10 * time.Millisecond) // Small delay for timestamp ordering
		}

		// Verify transaction history
		history, _ := h.transactionRepo.GetUserTransactions(testUser.ID, 20, 0)
		
		// Check count
		correctCount := len(history) >= len(transactions)
		
		// Check order (should be newest first)
		orderedCorrectly := true
		for i := 1; i < len(history); i++ {
			if history[i-1].CreatedAt.Before(history[i].CreatedAt) {
				orderedCorrectly = false
				break
			}
		}

		// Calculate final expected holding
		netShares := 0
		for _, tx := range transactions {
			if tx.action == "buy" {
				netShares += tx.quantity
			} else {
				netShares -= tx.quantity
			}
		}

		finalHolding, _ := h.portfolioRepo.GetUserStockHolding(testUser.ID, stock.ID)
		finalQuantityCorrect := finalHolding != nil && finalHolding.Quantity == netShares

		return map[string]interface{}{
			"transactions_executed":     len(transactions),
			"transactions_recorded":     len(history),
			"correct_count":            correctCount,
			"ordered_correctly":        orderedCorrectly,
			"expected_final_shares":    netShares,
			"actual_final_shares":      finalHolding.Quantity,
			"final_quantity_correct":   finalQuantityCorrect,
		}, nil
	})

	// Test 7: Concurrent Trading Conflicts
	h.runPortfolioTest(&suite, "Concurrent Trading Simulation", func() (map[string]interface{}, error) {
		testUser, err := h.createTestUser("concurrent_user", 20000)
		if err != nil {
			return nil, err
		}
		defer h.cleanupTestUser(testUser.ID)

		stock, err := h.getHealthyStock()
		if err != nil {
			return nil, err
		}

		// First buy some shares to enable selling
		err = h.marketService.BuyStock(testUser.ID, stock.ID, 20)
		if err != nil {
			return nil, err
		}

		// Simulate rapid trading (sequential for now, could be made concurrent)
		operations := []struct {
			action   string
			quantity int
		}{
			{"buy", 5},
			{"sell", 3},
			{"buy", 2},
			{"sell", 1},
			{"buy", 4},
		}

		successCount := 0
		for _, op := range operations {
			var err error
			if op.action == "buy" {
				err = h.marketService.BuyStock(testUser.ID, stock.ID, op.quantity)
			} else {
				err = h.marketService.SellStock(testUser.ID, stock.ID, op.quantity)
			}
			if err == nil {
				successCount++
			}
		}

		// Verify final state consistency
		finalHolding, _ := h.portfolioRepo.GetUserStockHolding(testUser.ID, stock.ID)
		finalUser, _ := h.userRepo.GetUserByID(testUser.ID)
		transactions, _ := h.transactionRepo.GetUserTransactions(testUser.ID, 50, 0)

		return map[string]interface{}{
			"operations_attempted":   len(operations),
			"operations_successful":  successCount,
			"final_shares":           finalHolding.Quantity,
			"final_cash":             finalUser.CashBalance,
			"transactions_recorded":  len(transactions),
			"data_consistency_ok":    successCount == len(operations), // All should succeed in this test
		}, nil
	})

	// Complete the suite
	suite.EndTime = time.Now()
	suite.TotalTests = len(suite.Tests)
	for _, test := range suite.Tests {
		if test.Status == "passed" {
			suite.Passed++
		} else {
			suite.Failed++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suite)
}

// Helper functions
func (h *PortfolioTestHandler) runPortfolioTest(suite *TestSuite, testName string, testFunc func() (map[string]interface{}, error)) {
	startTime := time.Now()
	
	result := TestResult{
		TestName: testName,
		Status:   "running",
	}
	
	details, err := testFunc()
	
	duration := time.Since(startTime)
	result.Duration = int(duration.Milliseconds())
	
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Test failed: %v", err)
	} else {
		result.Status = "passed"
		result.Message = "Test completed successfully"
		result.Details = details
	}
	
	suite.Tests = append(suite.Tests, result)
	
	log.Printf("Portfolio Test '%s': %s (duration: %d ms)", testName, result.Status, result.Duration)
}

func (h *PortfolioTestHandler) createTestUser(username string, cashBalance float64) (*models.User, error) {
	// Create unique username with timestamp
	uniqueUsername := fmt.Sprintf("%s_%d", username, time.Now().UnixNano())
	
	user, err := h.userRepo.CreateUser(uniqueUsername, "test123")
	if err != nil {
		return nil, err
	}
	
	// Update the user's cash balance
	err = h.userRepo.UpdateUserBalance(user.ID, cashBalance)
	if err != nil {
		// Clean up the created user on error
		h.userRepo.DeleteUser(user.ID)
		return nil, err
	}
	
	// Reload user with updated balance
	user, err = h.userRepo.GetUserByID(user.ID)
	if err != nil {
		return nil, err
	}
	
	log.Printf("Created test user: %s (ID: %d) with $%.2f", uniqueUsername, user.ID, cashBalance)
	return user, nil
}

func (h *PortfolioTestHandler) cleanupTestUser(userID int) {
	// Remove user's portfolio holdings
	portfolio, _ := h.portfolioRepo.GetUserPortfolio(userID)
	for _, holding := range portfolio {
		h.portfolioRepo.RemoveStockFromPortfolio(holding.ID)
	}
	
	// Delete user (this should cascade delete transactions)
	h.userRepo.DeleteUser(userID)
	log.Printf("Cleaned up test user ID: %d", userID)
}

func (h *PortfolioTestHandler) getHealthyStock() (*models.Stock, error) {
	stocks, err := h.stockRepo.GetAllStocks()
	if err != nil {
		return nil, err
	}
	
	// Find a healthy stock (active, reasonable price)
	for _, stock := range stocks {
		if stock.Status == models.StockActive && stock.CurrentPrice > 1 && stock.CurrentPrice < 1000 {
			return stock, nil
		}
	}
	
	return nil, fmt.Errorf("no suitable healthy stock found")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}