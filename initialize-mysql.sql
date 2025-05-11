-- Initialize the database schema for OfficeStonks

-- Users table with authentication information
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    cash_balance DECIMAL(15, 2) DEFAULT 10000.00,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Stocks table with current market information
CREATE TABLE IF NOT EXISTS stocks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(10) UNIQUE NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    current_price DECIMAL(15, 2) NOT NULL,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Portfolio holdings for each user
CREATE TABLE IF NOT EXISTS portfolio (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    stock_id INT NOT NULL,
    shares INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (stock_id) REFERENCES stocks(id),
    UNIQUE KEY user_stock (user_id, stock_id)
);

-- Transaction history for all trades
CREATE TABLE IF NOT EXISTS transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    stock_id INT NOT NULL,
    type ENUM('buy', 'sell') NOT NULL,
    shares INT NOT NULL,
    price_per_share DECIMAL(15, 2) NOT NULL,
    total_amount DECIMAL(15, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (stock_id) REFERENCES stocks(id)
);

-- Chat messages for the community feature
CREATE TABLE IF NOT EXISTS chat_messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Insert initial stocks
INSERT IGNORE INTO stocks (symbol, company_name, current_price) VALUES 
('AAPL', 'Apple Inc.', 150.00),
('MSFT', 'Microsoft Corporation', 300.00),
('AMZN', 'Amazon.com Inc.', 3300.00),
('GOOGL', 'Alphabet Inc.', 2800.00),
('META', 'Meta Platforms Inc.', 330.00),
('TSLA', 'Tesla Inc.', 700.00),
('NFLX', 'Netflix Inc.', 550.00),
('DIS', 'The Walt Disney Company', 175.00),
('NVDA', 'NVIDIA Corporation', 220.00),
('PYPL', 'PayPal Holdings Inc.', 240.00);

-- Create admin user if it doesn't exist
INSERT IGNORE INTO users (username, password_hash, cash_balance, is_admin)
VALUES ('admin', '$2a$10$JdvU7xXL6eLAC1ped9bY5.RMRxgNUT1Dg.Bh3ZJxXmVvIyAOKHYQu', 10000.00, TRUE);

-- Display created tables
SHOW TABLES;