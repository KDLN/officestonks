package repository

import (
	"log"
	"os"
	"path/filepath"
)

// Database schema SQL - embedded directly to avoid filesystem dependencies
const schemaSQL = `
-- Database schema for Office Stonks

-- Users Table
CREATE TABLE IF NOT EXISTS users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  cash_balance DECIMAL(15,2) DEFAULT 10000.00,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Stocks Table
CREATE TABLE IF NOT EXISTS stocks (
  id INT PRIMARY KEY AUTO_INCREMENT,
  symbol VARCHAR(10) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  sector VARCHAR(50),
  current_price DECIMAL(10,2) NOT NULL,
  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User Portfolios Table
CREATE TABLE IF NOT EXISTS portfolios (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  stock_id INT NOT NULL,
  quantity INT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (stock_id) REFERENCES stocks(id),
  UNIQUE KEY unique_user_stock (user_id, stock_id)
);

-- Transactions Table
CREATE TABLE IF NOT EXISTS transactions (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  stock_id INT NOT NULL,
  quantity INT NOT NULL,
  price DECIMAL(10,2) NOT NULL,
  transaction_type ENUM('buy', 'sell') NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (stock_id) REFERENCES stocks(id)
);

-- Chat Messages Table
CREATE TABLE IF NOT EXISTS chat_messages (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  message TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);
`

// Initial seed data SQL
const seedDataSQL = `
-- Initial seed data for stocks
INSERT IGNORE INTO stocks (symbol, name, sector, current_price) VALUES
('APPL', 'Apple Inc.', 'Technology', 150.00),
('GOOG', 'Alphabet Inc.', 'Technology', 2800.00),
('AMZN', 'Amazon.com Inc.', 'Technology', 3400.00),
('MSFT', 'Microsoft Corporation', 'Technology', 310.00),
('TSLA', 'Tesla, Inc.', 'Automotive', 950.00),
('JPM', 'JPMorgan Chase & Co.', 'Financial Services', 165.00),
('WMT', 'Walmart Inc.', 'Retail', 145.00),
('DIS', 'The Walt Disney Company', 'Entertainment', 185.00),
('NFLX', 'Netflix, Inc.', 'Entertainment', 580.00),
('PFE', 'Pfizer Inc.', 'Healthcare', 42.00);
`

// InitSchema initializes the database schema
func InitSchema() error {
	log.Println("Initializing database schema...")
	
	// Execute schema SQL
	_, err := DB.Exec(schemaSQL)
	if err != nil {
		log.Printf("Error executing schema SQL: %v", err)
		return err
	}
	
	log.Println("Database schema created successfully.")
	
	// Check if stocks table has data
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM stocks").Scan(&count)
	if err != nil {
		log.Printf("Error checking stocks count: %v", err)
		return err
	}
	
	// Only seed data if no stocks exist
	if count == 0 {
		log.Println("Seeding initial stock data...")
		_, err = DB.Exec(seedDataSQL)
		if err != nil {
			log.Printf("Error seeding data: %v", err)
			return err
		}
		log.Println("Initial seed data loaded successfully.")
	} else {
		log.Printf("Stocks table already has %d records, skipping seed data.", count)
	}
	
	return nil
}