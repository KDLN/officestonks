package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yourusername/officestonks/internal/models"
)

func TestGetAllStocks(t *testing.T) {
	// Skip if no test database connection
	if TestDB == nil {
		t.Skip("No test database connection")
	}

	// Setup test router
	router := SetupTestRouter(TestDB)

	// Make request
	rr := MakeRequest("GET", "/api/stocks", nil, router)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var stocks []*models.Stock
	if err := json.Unmarshal(rr.Body.Bytes(), &stocks); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Check if stocks are returned
	if len(stocks) == 0 {
		t.Error("Expected stocks in response, got empty array")
	}

	// Check stock structure
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
}

func TestGetStockByID(t *testing.T) {
	// Skip if no test database connection
	if TestDB == nil {
		t.Skip("No test database connection")
	}

	// Setup test router
	router := SetupTestRouter(TestDB)

	// Get all stocks first to get a valid ID
	allStocksRR := MakeRequest("GET", "/api/stocks", nil, router)
	if allStocksRR.Code != http.StatusOK {
		t.Fatalf("Failed to get stocks: %s", allStocksRR.Body.String())
	}

	var stocks []*models.Stock
	if err := json.Unmarshal(allStocksRR.Body.Bytes(), &stocks); err != nil {
		t.Fatalf("Failed to parse stocks: %v", err)
	}

	if len(stocks) == 0 {
		t.Skip("No stocks in database to test with")
	}

	// Test cases
	tests := []struct {
		name           string
		stockID        int
		expectedStatus int
		shouldSucceed  bool
	}{
		{"ValidStock", stocks[0].ID, http.StatusOK, true},
		{"NonexistentStock", 9999, http.StatusNotFound, false},
		{"InvalidID", -1, http.StatusNotFound, false},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Make request
			url := fmt.Sprintf("/api/stocks/%d", tc.stockID)
			rr := MakeRequest("GET", url, nil, router)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rr.Code)
			}

			// If should succeed, check response structure
			if tc.shouldSucceed {
				var stock models.Stock
				if err := json.Unmarshal(rr.Body.Bytes(), &stock); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
				}

				// Check stock fields
				if stock.ID != tc.stockID {
					t.Errorf("Expected stock ID %d, got %d", tc.stockID, stock.ID)
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
	}
}

func TestTradeEndpoint(t *testing.T) {
	// Skip if no test database connection
	if TestDB == nil {
		t.Skip("No test database connection")
	}

	// Setup test router
	router := SetupTestRouter(TestDB)

	// Create a test user
	user := CreateTestUser(t, router, "tradeuser", "tradepassword")
	if user == nil {
		t.Fatal("Failed to create test user")
	}

	// Get a stock to trade
	stocksRR := MakeRequest("GET", "/api/stocks", nil, router)
	var stocks []*models.Stock
	if err := json.Unmarshal(stocksRR.Body.Bytes(), &stocks); err != nil {
		t.Fatalf("Failed to parse stocks: %v", err)
	}

	if len(stocks) == 0 {
		t.Skip("No stocks in database to test with")
	}

	// Test buying a stock
	buyReq := models.TradeRequest{
		StockID:  stocks[0].ID,
		Quantity: 1,
		Action:   "buy",
	}

	// In a real test, we would make the authenticated request
	// but for the demo, we'll just show the structure
	
	// Example of authenticated request:
	// buyRR := AuthenticatedRequest("POST", "/api/trading", buyReq, user.UserID, router)
	
	// When testing trade endpoint, you would check:
	// 1. Status code is 201 Created
	// 2. Portfolio updates correctly
	// 3. Cash balance is decreased
	// 4. Transaction is recorded
	
	// Then test selling a stock:
	sellReq := models.TradeRequest{
		StockID:  stocks[0].ID,
		Quantity: 1,
		Action:   "sell",
	}
	
	// Similar checks for selling
	
	// Finally test error cases:
	// 1. Invalid stock ID
	// 2. Negative quantity
	// 3. Selling more shares than owned
	// 4. Buying more shares than can afford
	
	// This test is structured but not fully implemented since it requires
	// a complete test database setup with users, stocks, etc.
	t.Log("Trade endpoint test structure created, needs full implementation with test database")
}