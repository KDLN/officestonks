-- Stock Lifecycle System Updates
-- This migration adds support for dynamic stock creation, IPOs, and lifecycle management

-- Add new columns to stocks table
ALTER TABLE stocks 
ADD COLUMN launch_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN initial_price DECIMAL(10, 2) DEFAULT NULL,
ADD COLUMN market_cap_category ENUM('penny', 'small', 'mid', 'large') DEFAULT 'mid',
ADD COLUMN volatility_profile ENUM('stable', 'normal', 'volatile', 'extreme') DEFAULT 'normal',
ADD COLUMN company_description TEXT DEFAULT NULL,
ADD COLUMN ipo_shares_available INT DEFAULT 1000000,
ADD COLUMN total_shares INT DEFAULT 10000000,
ADD COLUMN founded_year INT DEFAULT NULL;

-- Create stock lifecycle events table
CREATE TABLE IF NOT EXISTS stock_lifecycle_events (
    id INT PRIMARY KEY AUTO_INCREMENT,
    stock_id INT NOT NULL,
    event_type ENUM('ipo', 'delisting', 'bankruptcy', 'merger', 'split', 'warning') NOT NULL,
    event_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    details JSON,
    price_at_event DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_stock_id (stock_id),
    INDEX idx_event_date (event_date),
    FOREIGN KEY (stock_id) REFERENCES stocks(id) ON DELETE CASCADE
);

-- Create stock templates table for quick generation
CREATE TABLE IF NOT EXISTS stock_templates (
    id INT PRIMARY KEY AUTO_INCREMENT,
    template_name VARCHAR(100) NOT NULL,
    company_name_pattern VARCHAR(200) NOT NULL,
    sector_id INT,
    min_initial_price DECIMAL(10, 2) DEFAULT 0.10,
    max_initial_price DECIMAL(10, 2) DEFAULT 100.00,
    volatility_profile ENUM('stable', 'normal', 'volatile', 'extreme') DEFAULT 'normal',
    description_template TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sector_id) REFERENCES sectors(id) ON DELETE SET NULL
);

-- Create IPO queue table
CREATE TABLE IF NOT EXISTS ipo_queue (
    id INT PRIMARY KEY AUTO_INCREMENT,
    stock_symbol VARCHAR(10) UNIQUE NOT NULL,
    company_name VARCHAR(100) NOT NULL,
    sector_id INT,
    initial_price DECIMAL(10, 2) NOT NULL,
    shares_available INT DEFAULT 1000000,
    launch_date TIMESTAMP NOT NULL,
    status ENUM('pending', 'launched', 'cancelled') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_launch_date (launch_date),
    INDEX idx_status (status),
    FOREIGN KEY (sector_id) REFERENCES sectors(id) ON DELETE SET NULL
);

-- Create market events table
CREATE TABLE IF NOT EXISTS market_events (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_name VARCHAR(100) NOT NULL,
    event_type ENUM('sector_crash', 'sector_boom', 'market_crash', 'market_rally', 'ipo_wave') NOT NULL,
    sector_id INT DEFAULT NULL,
    impact_percentage DECIMAL(5, 2) NOT NULL,
    duration_minutes INT DEFAULT 60,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP NULL,
    created_by INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_active_events (start_time, end_time),
    FOREIGN KEY (sector_id) REFERENCES sectors(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Add some initial stock templates
INSERT INTO stock_templates (template_name, company_name_pattern, sector_id, min_initial_price, max_initial_price, volatility_profile, description_template) VALUES
('Tech Startup', '{adjective} {noun} Technologies', 1, 0.10, 5.00, 'extreme', 'A cutting-edge technology startup focusing on {technology} solutions.'),
('Penny Mining', '{location} {mineral} Mining Corp', 3, 0.05, 2.00, 'volatile', 'A small mining operation exploring {mineral} deposits in {location}.'),
('Biotech Venture', '{prefix}Gen Therapeutics', 5, 0.50, 10.00, 'extreme', 'A biotech company developing innovative {medical} treatments.'),
('Entertainment Startup', '{adjective} {entertainment} Studios', 4, 1.00, 20.00, 'volatile', 'An entertainment company specializing in {entertainment} content.'),
('Green Energy', '{region} {energy} Power', 2, 5.00, 50.00, 'normal', 'A renewable energy company harnessing {energy} power in {region}.');

-- Add indexes for performance
CREATE INDEX idx_stocks_market_cap ON stocks(market_cap_category);
CREATE INDEX idx_stocks_volatility ON stocks(volatility_profile);
CREATE INDEX idx_stocks_status_price ON stocks(status, current_price);