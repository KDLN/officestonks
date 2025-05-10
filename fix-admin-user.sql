-- COMPLETE DATABASE INITIALIZATION SCRIPT FOR OFFICESTONKS
-- This script will:
-- 1. Fix admin user issues
-- 2. Reset stock prices
-- 3. Clear chat messages

-- =================== USER FIXES ===================
-- First make sure the is_admin column exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;

-- Show existing users
SELECT 'Existing users before changes:' AS Message;
SELECT id, username, is_admin, cash_balance FROM users;

-- Create admin user if it doesn't exist
INSERT INTO users (username, password_hash, cash_balance, is_admin)
SELECT 'admin', '$2a$10$JdvU7xXL6eLAC1ped9bY5.RMRxgNUT1Dg.Bh3ZJxXmVvIyAOKHYQu', 10000.00, TRUE
FROM dual
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');

-- Update existing admin user to have admin privileges
UPDATE users SET is_admin = TRUE WHERE username = 'admin';

-- Make sure admin has cash
UPDATE users SET cash_balance = 10000.00 WHERE username = 'admin' AND cash_balance <= 0;

-- =================== STOCK FIXES ===================
-- Make sure stocks table exists
CREATE TABLE IF NOT EXISTS stocks (
  id INT PRIMARY KEY AUTO_INCREMENT,
  symbol VARCHAR(10) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  sector VARCHAR(50),
  current_price DECIMAL(10,2) NOT NULL,
  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Show existing stocks
SELECT 'Existing stocks before changes:' AS Message;
SELECT id, symbol, name, current_price FROM stocks;

-- Only add stocks if none exist
INSERT IGNORE INTO stocks (symbol, name, sector, current_price)
VALUES
  ('AAPL', 'Apple Inc.', 'Technology', 150.00),
  ('MSFT', 'Microsoft Corporation', 'Technology', 250.00),
  ('GOOGL', 'Alphabet Inc.', 'Technology', 2500.00),
  ('AMZN', 'Amazon.com Inc.', 'Technology', 3000.00),
  ('META', 'Meta Platforms Inc.', 'Technology', 300.00),
  ('TSLA', 'Tesla, Inc.', 'Automotive', 700.00),
  ('NFLX', 'Netflix, Inc.', 'Entertainment', 550.00);

-- Reset all stock prices to new random values
UPDATE stocks
SET current_price = ROUND(RAND() * 1000, 2),
    last_updated = NOW();

-- =================== CHAT FIXES ===================
-- Make sure chat_messages table exists
CREATE TABLE IF NOT EXISTS chat_messages (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  message TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Delete all chat messages
DELETE FROM chat_messages;

-- Add a test message
INSERT INTO chat_messages (user_id, message)
SELECT id, 'Admin message: Chat has been cleared and reset.'
FROM users
WHERE username = 'admin'
LIMIT 1;

-- =================== PORTFOLIO FIXES ===================
-- Make sure portfolios table exists
CREATE TABLE IF NOT EXISTS portfolios (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  stock_id INT NOT NULL,
  quantity INT NOT NULL,
  UNIQUE KEY unique_user_stock (user_id, stock_id)
);

-- Make sure transactions table exists
CREATE TABLE IF NOT EXISTS transactions (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  stock_id INT NOT NULL,
  quantity INT NOT NULL,
  price DECIMAL(10,2) NOT NULL,
  transaction_type ENUM('buy', 'sell') NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =================== FINAL VERIFICATION ===================
-- Show the final state of the database
SELECT 'FINAL USERS:' AS Message;
SELECT id, username, is_admin, cash_balance FROM users;

SELECT 'FINAL STOCKS:' AS Message;
SELECT id, symbol, name, current_price FROM stocks;

SELECT 'CHAT MESSAGES:' AS Message;
SELECT COUNT(*) AS message_count FROM chat_messages;