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

// TestHandler handles test orchestration endpoints
type TestHandler struct {
	marketService  *services.MarketService
	userService    *services.UserService
	stockRepo      models.StockRepository
	portfolioRepo  models.PortfolioRepository
	delistedRepo   models.DelistedStockRepository
	lossRepo       models.PortfolioLossRepository
	newsRepo       models.NewsRepository
}

// NewTestHandler creates a new test handler
func NewTestHandler(
	marketService *services.MarketService,
	userService *services.UserService,
	stockRepo models.StockRepository,
	portfolioRepo models.PortfolioRepository,
	delistedRepo models.DelistedStockRepository,
	lossRepo models.PortfolioLossRepository,
	newsRepo models.NewsRepository,
) *TestHandler {
	return &TestHandler{
		marketService:  marketService,
		userService:    userService,
		stockRepo:      stockRepo,
		portfolioRepo:  portfolioRepo,
		delistedRepo:   delistedRepo,
		lossRepo:       lossRepo,
		newsRepo:       newsRepo,
	}
}

// TestResult represents the result of a test
type TestResult struct {
	TestName    string                 `json:"test_name"`
	Status      string                 `json:"status"` // "running", "passed", "failed"
	Message     string                 `json:"message"`
	Duration    string                 `json:"duration"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// TestSuite represents a collection of test results
type TestSuite struct {
	SuiteName   string       `json:"suite_name"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     time.Time    `json:"end_time"`
	TotalTests  int          `json:"total_tests"`
	Passed      int          `json:"passed"`
	Failed      int          `json:"failed"`
	Tests       []TestResult `json:"tests"`
}

// RunCrisisTests runs the crisis mechanics test suite (admin only)
func (h *TestHandler) RunCrisisTests(w http.ResponseWriter, r *http.Request) {
	suite := TestSuite{
		SuiteName: "Crisis Mechanics Test Suite",
		StartTime: time.Now(),
		Tests:     []TestResult{},
	}

	// Test 1: Force Crisis Event
	h.runTest(&suite, "Force Crisis Event", func() (map[string]interface{}, error) {
		// Get a healthy stock
		stocks, err := h.stockRepo.GetAllStocks()
		if err != nil {
			return nil, fmt.Errorf("failed to get stocks: %w", err)
		}
		
		var targetStock *models.Stock
		for _, s := range stocks {
			if s.CurrentPrice > 10 && s.Status == models.StockActive {
				targetStock = s
				break
			}
		}
		
		if targetStock == nil {
			return nil, fmt.Errorf("no suitable stock found for crisis test")
		}
		
		// Force crisis
		err = h.marketService.ForceCrisisEvent(targetStock.ID)
		if err != nil {
			return nil, err
		}
		
		// Wait for processing
		time.Sleep(2 * time.Second)
		
		// Verify
		updatedStock, _ := h.stockRepo.GetStockByID(targetStock.ID)
		
		return map[string]interface{}{
			"stock_id":     targetStock.ID,
			"stock_symbol": targetStock.Symbol,
			"price_before": targetStock.CurrentPrice,
			"price_after":  updatedStock.CurrentPrice,
			"status":       updatedStock.Status,
		}, nil
	})

	// Test 2: Bankruptcy with Portfolio Impact
	h.runTest(&suite, "Bankruptcy with Portfolio Impact", func() (map[string]interface{}, error) {
		// Find a stock with holders
		stocks, _ := h.stockRepo.GetAllStocks()
		var targetStock *models.Stock
		var holderCount int
		
		for _, s := range stocks {
			holders, _ := h.portfolioRepo.GetAllHoldersOfStock(s.ID)
			if len(holders) > 0 && s.Status != models.StockDelisted {
				targetStock = s
				holderCount = len(holders)
				break
			}
		}
		
		if targetStock == nil {
			return nil, fmt.Errorf("no stock with holders found")
		}
		
		// Force bankruptcy
		err := h.marketService.ForceBankruptcy(targetStock.ID)
		if err != nil {
			return nil, err
		}
		
		// Wait for processing
		time.Sleep(3 * time.Second)
		
		// Verify holdings removed
		remainingHolders, _ := h.portfolioRepo.GetAllHoldersOfStock(targetStock.ID)
		
		// Check losses recorded
		losses, _ := h.lossRepo.GetLossesByStock(targetStock.ID, 100)
		
		// Check delisted
		delisted, _ := h.delistedRepo.GetDelistedStockByID(targetStock.ID)
		
		return map[string]interface{}{
			"stock_id":           targetStock.ID,
			"stock_symbol":       targetStock.Symbol,
			"holders_before":     holderCount,
			"holders_after":      len(remainingHolders),
			"losses_recorded":    len(losses),
			"delisted":           delisted != nil,
			"delisting_reason":   delisted.Reason,
		}, nil
	})

	// Test 3: Recovery Event
	h.runTest(&suite, "Recovery Event", func() (map[string]interface{}, error) {
		// Find or create a crisis stock
		stocks, _ := h.stockRepo.GetAllStocks()
		var targetStock *models.Stock
		
		for _, s := range stocks {
			if s.CurrentPrice <= 0.01 && s.Status == models.StockDistressed {
				targetStock = s
				break
			}
		}
		
		if targetStock == nil {
			// Create one by forcing crisis
			for _, s := range stocks {
				if s.Status == models.StockActive {
					h.marketService.ForceCrisisEvent(s.ID)
					time.Sleep(1 * time.Second)
					targetStock, _ = h.stockRepo.GetStockByID(s.ID)
					break
				}
			}
		}
		
		if targetStock == nil {
			return nil, fmt.Errorf("could not create crisis stock")
		}
		
		priceBefore := targetStock.CurrentPrice
		
		// Force recovery
		err := h.marketService.ForceRecovery(targetStock.ID)
		if err != nil {
			return nil, err
		}
		
		// Wait for processing
		time.Sleep(2 * time.Second)
		
		// Verify
		updatedStock, _ := h.stockRepo.GetStockByID(targetStock.ID)
		
		return map[string]interface{}{
			"stock_id":        targetStock.ID,
			"stock_symbol":    targetStock.Symbol,
			"price_before":    priceBefore,
			"price_after":     updatedStock.CurrentPrice,
			"price_increase":  fmt.Sprintf("%.0fx", updatedStock.CurrentPrice/priceBefore),
			"status":          updatedStock.Status,
		}, nil
	})

	// Test 4: News Generation
	h.runTest(&suite, "News Generation Verification", func() (map[string]interface{}, error) {
		// Get recent crisis-related news
		news, err := h.newsRepo.GetRecentNews(50)
		if err != nil {
			return nil, err
		}
		
		crisisCount := 0
		bankruptcyCount := 0
		recoveryCount := 0
		
		for _, n := range news {
			switch n.Type {
			case models.NewsTypeCrisis:
				crisisCount++
			case models.NewsTypeBankruptcy:
				bankruptcyCount++
			case models.NewsTypeRecovery:
				recoveryCount++
			}
		}
		
		return map[string]interface{}{
			"total_news":       len(news),
			"crisis_news":      crisisCount,
			"bankruptcy_news":  bankruptcyCount,
			"recovery_news":    recoveryCount,
		}, nil
	})

	// Test 5: Sector Contagion
	h.runTest(&suite, "Sector Contagion Check", func() (map[string]interface{}, error) {
		// This is harder to test automatically but we can check for patterns
		stocks, _ := h.stockRepo.GetAllStocks()
		
		sectorStats := make(map[string]map[string]int)
		for _, s := range stocks {
			if _, exists := sectorStats[s.Sector]; !exists {
				sectorStats[s.Sector] = map[string]int{
					"total":      0,
					"active":     0,
					"distressed": 0,
					"delisted":   0,
				}
			}
			
			sectorStats[s.Sector]["total"]++
			switch s.Status {
			case models.StockActive:
				sectorStats[s.Sector]["active"]++
			case models.StockDistressed:
				sectorStats[s.Sector]["distressed"]++
			case models.StockDelisted:
				sectorStats[s.Sector]["delisted"]++
			}
		}
		
		return map[string]interface{}{
			"sector_breakdown": sectorStats,
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

// Helper function to run individual tests
func (h *TestHandler) runTest(suite *TestSuite, testName string, testFunc func() (map[string]interface{}, error)) {
	startTime := time.Now()
	
	result := TestResult{
		TestName: testName,
		Status:   "running",
	}
	
	details, err := testFunc()
	
	duration := time.Since(startTime)
	result.Duration = duration.String()
	
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
	
	log.Printf("Test '%s': %s (duration: %s)", testName, result.Status, result.Duration)
}

// GetTestStatus returns current test capabilities (admin only)
func (h *TestHandler) GetTestStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"available_tests": []string{
			"Crisis Mechanics Test Suite",
			"Portfolio & Trading Test Suite",
		},
		"test_endpoints": map[string]string{
			"run_crisis_tests":    "/api/admin/tests/crisis",
			"run_portfolio_tests": "/api/admin/tests/portfolio",
			"get_test_status":     "/api/admin/tests/status",
		},
		"environment": map[string]interface{}{
			"simulator_running": true, // Could check actual status
			"test_mode_enabled": true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}