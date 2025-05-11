# MySQL Deployment for OfficeStonks

## Overview

OfficeStonks requires a MySQL database. When deploying on Railway, you need to set up both the main application service and a MySQL database service.

## Steps to Deploy MySQL on Railway

1. Create a new MySQL service in your Railway project:
   - Go to your Railway project
   - Click "New Service"
   - Select "MySQL"
   - Railway will automatically provision a MySQL instance

2. Configure Environment Variables:
   - Railway will automatically generate environment variables like:
     - `MYSQLHOST`
     - `MYSQLPORT`
     - `MYSQLUSER`
     - `MYSQLPASSWORD`
     - `MYSQLDATABASE`
   - These variables will be used by the main application service

3. Initialize the Database:
   - Railway MySQL setup doesn't require a Dockerfile
   - You can use the "Raw SQL" option in the Railway dashboard to run the schema initialization SQL

## Schema Initialization

Copy and paste the following SQL into Railway's "Raw SQL" interface for the MySQL service:

```sql
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
```

## Linking to Main Application

After setting up the MySQL service, make sure your main application service has access to the database environment variables. Railway should automatically share these variables between services in the same project.

## Testing the Connection

You can verify the connection from your main application to the MySQL database by:

1. Accessing the logs of your main application service
2. Checking for successful database connection messages