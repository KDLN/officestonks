package tests

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

// TestMain is the entry point for running all tests in this package
func TestMain(m *testing.M) {
	// Setup test environment
	if err := setupTests(); err != nil {
		log.Printf("Failed to setup tests: %v", err)
		os.Exit(1)
	}

	// Run tests
	exitCode := m.Run()

	// Cleanup test environment
	if err := cleanupTests(); err != nil {
		log.Printf("Failed to cleanup tests: %v", err)
	}

	// Exit with test result code
	os.Exit(exitCode)
}

// setupTests initializes the test environment
func setupTests() error {
	// Initialize test database connection with retries
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		TestDB, err = SetupTestDB()
		if err == nil {
			break
		}
		log.Printf("Attempt %d: Failed to connect to test database: %v", i+1, err)
		if i < maxRetries-1 {
			// Wait before retrying
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		log.Printf("Warning: Could not connect to test database after %d attempts: %v", maxRetries, err)
		log.Println("Tests requiring database will be skipped")
		return nil
	}

	// Initialize test data
	if err := setupTestData(); err != nil {
		return fmt.Errorf("failed to setup test data: %w", err)
	}

	return nil
}

// setupTestData initializes test data in the database
func setupTestData() error {
	// Only proceed if we have a database connection
	if TestDB == nil {
		return nil
	}

	// Clear existing test data
	if err := clearTestData(); err != nil {
		return err
	}

	// Create test stocks
	if err := createTestStocks(); err != nil {
		return err
	}

	return nil
}

// clearTestData removes existing test data
func clearTestData() error {
	// Disable foreign key checks temporarily
	_, err := TestDB.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		return err
	}

	// Truncate tables
	tables := []string{"transactions", "portfolios", "users", "stocks"}
	for _, table := range tables {
		_, err := TestDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table))
		if err != nil {
			return err
		}
	}

	// Re-enable foreign key checks
	_, err = TestDB.Exec("SET FOREIGN_KEY_CHECKS = 1")
	return err
}

// createTestStocks adds test stocks to the database
func createTestStocks() error {
	// Add test stocks
	stocks := []struct {
		symbol  string
		name    string
		sector  string
		price   float64
	}{
		{"AAPL", "Apple Inc.", "Technology", 150.0},
		{"MSFT", "Microsoft Corporation", "Technology", 300.0},
		{"GOOG", "Alphabet Inc.", "Technology", 2800.0},
		{"AMZN", "Amazon.com Inc.", "Technology", 3400.0},
		{"TSLA", "Tesla, Inc.", "Automotive", 950.0},
	}

	for _, stock := range stocks {
		_, err := TestDB.Exec(
			"INSERT INTO stocks (symbol, name, sector, current_price) VALUES (?, ?, ?, ?)",
			stock.symbol, stock.name, stock.sector, stock.price,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanupTests cleans up the test environment
func cleanupTests() error {
	// Close database connection if open
	if TestDB != nil {
		return TestDB.Close()
	}
	return nil
}

// TestDatabaseConnection verifies the test database connection
func TestDatabaseConnection(t *testing.T) {
	if TestDB == nil {
		t.Skip("No test database connection")
	}

	// Ping database to verify connection
	if err := TestDB.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Check if we have test stocks
	var count int
	err := TestDB.QueryRow("SELECT COUNT(*) FROM stocks").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count stocks: %v", err)
	}

	if count == 0 {
		t.Fatal("No test stocks found in database")
	}

	t.Logf("Successfully connected to test database with %d test stocks", count)
}