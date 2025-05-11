-- Initialize admin user and make sure the is_admin column exists

-- First check if the is_admin column exists in the users table
SET @columnExists = 0;
SELECT COUNT(*) INTO @columnExists 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = 'railway' 
AND TABLE_NAME = 'users' 
AND COLUMN_NAME = 'is_admin';

-- If column doesn't exist, add it
SET @alterQuery = IF(@columnExists = 0, 
                    'ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;',
                    'SELECT "is_admin column already exists" AS Message;');
PREPARE stmt FROM @alterQuery;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Make sure tables exist
CREATE TABLE IF NOT EXISTS users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  cash_balance DECIMAL(15,2) DEFAULT 10000.00,
  is_admin BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stocks (
  id INT PRIMARY KEY AUTO_INCREMENT,
  symbol VARCHAR(10) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  sector VARCHAR(50),
  current_price DECIMAL(10,2) NOT NULL,
  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add admin user if it doesn't exist
INSERT INTO users (username, password_hash, cash_balance, is_admin)
SELECT 'admin', '$2a$10$l6jzERQJiOVnWw8FN2qQw.fxJfZsXnuKNtGV.OU63s8SLsBJBvvV2', 10000.00, 1
FROM dual
WHERE NOT EXISTS (SELECT * FROM users WHERE username = 'admin');

-- Add stock data if none exists
INSERT INTO stocks (symbol, name, sector, current_price)
SELECT 'AAPL', 'Apple Inc.', 'Technology', 150.00
FROM dual
WHERE NOT EXISTS (SELECT * FROM stocks WHERE symbol = 'AAPL');

INSERT INTO stocks (symbol, name, sector, current_price)
SELECT 'MSFT', 'Microsoft Corporation', 'Technology', 250.00
FROM dual
WHERE NOT EXISTS (SELECT * FROM stocks WHERE symbol = 'MSFT');

INSERT INTO stocks (symbol, name, sector, current_price)
SELECT 'GOOGL', 'Alphabet Inc.', 'Technology', 2500.00
FROM dual
WHERE NOT EXISTS (SELECT * FROM stocks WHERE symbol = 'GOOGL');

-- Show results
SELECT 'Users in database:' AS Message;
SELECT id, username, is_admin, cash_balance FROM users;

SELECT 'Stocks in database:' AS Message;
SELECT id, symbol, name, current_price FROM stocks;