package repository

import (
	"log"
)

// Database schema SQL - embedded directly to avoid filesystem dependencies
const schemaSQL = `
-- Database schema for Office Stonks

-- Users Table
CREATE TABLE IF NOT EXISTS users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  supabase_id VARCHAR(255) NULL UNIQUE,
  cash_balance DECIMAL(15,2) DEFAULT 10000.00,
  is_admin BOOLEAN DEFAULT FALSE,
  last_login TIMESTAMP NULL,
  login_count INT DEFAULT 0,
  total_trades INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Add supabase_id column if it doesn't exist (for existing databases)
-- MySQL doesn't support IF NOT EXISTS for ADD COLUMN, so we'll handle it in code

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

-- News Table
CREATE TABLE IF NOT EXISTS news (
  id INT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sectors Table
CREATE TABLE IF NOT EXISTS sectors (
  id INT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(50) UNIQUE NOT NULL,
  description TEXT,
  trend FLOAT DEFAULT 0,
  volatility_modifier FLOAT DEFAULT 1.0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Delisted Stocks Table (tracks stocks that went bankrupt)
CREATE TABLE IF NOT EXISTS delisted_stocks (
  id INT PRIMARY KEY,
  symbol VARCHAR(10) NOT NULL,
  name VARCHAR(100) NOT NULL,
  sector VARCHAR(50),
  final_price DECIMAL(10,2),
  delisted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  reason ENUM('bankruptcy', 'merger', 'admin') DEFAULT 'bankruptcy'
);

-- Portfolio Losses Table (tracks player losses from bankruptcies)
CREATE TABLE IF NOT EXISTS portfolio_losses (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  stock_id INT NOT NULL,
  stock_symbol VARCHAR(10) NOT NULL,
  stock_name VARCHAR(100) NOT NULL,
  quantity INT NOT NULL,
  loss_amount DECIMAL(15,2) NOT NULL,
  delisted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Changelog Table (tracks system updates and announcements)
CREATE TABLE IF NOT EXISTS changelog (
  id INT PRIMARY KEY AUTO_INCREMENT,
  version VARCHAR(20) NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  changes JSON,
  change_type ENUM('feature', 'improvement', 'bugfix', 'breaking') NOT NULL,
  is_major BOOLEAN DEFAULT FALSE,
  is_visible BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by INT NULL,
  FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Audit Logs Table
CREATE TABLE IF NOT EXISTS audit_logs (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  action VARCHAR(50) NOT NULL,
  ip_address VARCHAR(45),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

-- News Items Table (for crisis events and market news)
CREATE TABLE IF NOT EXISTS news_items (
  id INT PRIMARY KEY AUTO_INCREMENT,
  type ENUM('company', 'sector', 'market', 'crisis', 'recovery', 'bankruptcy') NOT NULL,
  stock_id INT NULL,
  stock_symbol VARCHAR(10) NULL,
  sector_name VARCHAR(50) NULL,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  impact_type ENUM('immediate', 'gradual') DEFAULT 'immediate',
  impact_score INT NOT NULL DEFAULT 0, -- -100 to +100
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP NOT NULL,
  is_automated BOOLEAN DEFAULT TRUE,
  FOREIGN KEY (stock_id) REFERENCES stocks(id) ON DELETE SET NULL
);

-- User Sessions Table (for monitoring)
CREATE TABLE IF NOT EXISTS user_sessions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    login_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    logout_time TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    trades_count INT DEFAULT 0,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_sessions_user_id (user_id),
    INDEX idx_user_sessions_active (is_active),
    INDEX idx_user_sessions_login_time (login_time)
);

-- User Activity Table (for monitoring)
CREATE TABLE IF NOT EXISTS user_activity (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    username VARCHAR(50) NOT NULL,
    session_id INT NULL,
    action VARCHAR(100) NOT NULL,
    details TEXT,
    ip_address VARCHAR(45) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    response_time_ms INT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES user_sessions(id) ON DELETE SET NULL,
    INDEX idx_user_activity_user_id (user_id),
    INDEX idx_user_activity_username (username),
    INDEX idx_user_activity_action (action),
    INDEX idx_user_activity_timestamp (timestamp),
    INDEX idx_user_activity_success (success)
);

-- System Metrics Table (for monitoring)
CREATE TABLE IF NOT EXISTS system_metrics (
    id INT AUTO_INCREMENT PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    active_users INT DEFAULT 0,
    total_sessions INT DEFAULT 0,
    active_sessions INT DEFAULT 0,
    trades_per_hour INT DEFAULT 0,
    websocket_connections INT DEFAULT 0,
    database_health ENUM('healthy', 'degraded', 'down') DEFAULT 'healthy',
    error_rate DECIMAL(5,4) DEFAULT 0.0000,
    avg_response_time_ms DECIMAL(8,2) DEFAULT 0.00,
    INDEX idx_system_metrics_timestamp (timestamp)
);
`

// Initial seed data SQL
const seedDataSQL = `
-- Initial seed data for sectors
INSERT IGNORE INTO sectors (name, description) VALUES
('Technology', 'Software, hardware, and internet companies'),
('Automotive', 'Vehicle manufacturers and suppliers'),
('Financial Services', 'Banks, insurance, and financial institutions'),
('Retail', 'Consumer retail and e-commerce'),
('Entertainment', 'Media, streaming, and entertainment companies'),
('Healthcare', 'Pharmaceuticals, biotech, and medical devices');

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

-- Add sector_id column to stocks table for foreign key relationship
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS sector_id INT NULL;
ALTER TABLE stocks ADD CONSTRAINT fk_stocks_sector FOREIGN KEY (sector_id) REFERENCES sectors(id) ON DELETE SET NULL;

-- Update existing stocks with sector_id based on sector name
UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Technology') WHERE sector = 'Technology';
UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Automotive') WHERE sector = 'Automotive';
UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Financial Services') WHERE sector = 'Financial Services';
UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Retail') WHERE sector = 'Retail';
UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Entertainment') WHERE sector = 'Entertainment';
UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Healthcare') WHERE sector = 'Healthcare';

-- Add default admin user (password is 'admin123')
INSERT IGNORE INTO users (username, password_hash, cash_balance, is_admin) VALUES
('admin', '$2a$10$l6jzERQJiOVnWw8FN2qQw.fxJfZsXnuKNtGV.OU63s8SLsBJBvvV2', 10000.00, 1);

-- Initial changelog entries
INSERT IGNORE INTO changelog (version, title, description, changes, change_type, is_major, is_visible) VALUES
('v1.0.0', 'Office Stonks Launch', 'Initial release of the multiplayer stock market simulation game.', 
 JSON_ARRAY(
   'Real-time stock trading with live price updates',
   'Portfolio management and transaction history',
   'Leaderboard rankings by portfolio value',
   'Live chat system for social interaction',
   'Admin controls for market management'
 ), 'feature', true, true),
('v1.1.0', 'Market Sectors Foundation', 'Introduced market sectors with correlated stock movements for more realistic trading.', 
 JSON_ARRAY(
   'Added 6 market sectors: Technology, Automotive, Financial Services, Retail, Entertainment, Healthcare',
   'Stock prices now influenced by both individual trends (70%) and sector trends (30%)',
   'Sector-wide correlations create realistic market behavior',
   'Enhanced market simulator with sector tracking',
   'Database schema updated to support sector relationships'
 ), 'feature', true, true),
('v1.2.0', 'Crisis & News System', 'Major update transforming crisis events into exciting high-stakes gameplay with comprehensive news coverage.', 
 JSON_ARRAY(
   'Price Zone Volatility: Penny stocks (10%), Low-cap (7%), Mid-cap (5%), Large-cap (3%) for realistic market behavior',
   'Breaking News Ticker: Auto-rotating crisis alerts with play/pause controls on dashboard',
   'Enhanced News Display: Filter by Crisis, Bankruptcy, Recovery, Sector with color-coded items and stock symbols',
   'Portfolio Crisis Alerts: Real-time warnings for stocks at $0.01 with bankruptcy risk and recovery potential',
   'Crisis Mechanics: 5% bankruptcy chance, 3% recovery chance every 2 seconds for $0.01 stocks',
   'Trade Frequency Limiting: 5-second cooldown with 20 trades/hour limit per user for security',
   'Database Integration: Sector foreign key relationships and complete schema for crisis tracking',
   'Mobile Responsive: All new components optimized for mobile devices with smooth animations'
 ), 'feature', true, true);
`

// InitSchema initializes the database schema
func InitSchema() error {
	log.Println("Initializing database schema...")

	// Execute schema SQL with retry
	_, err := RetryExec(DB, schemaSQL)
	if err != nil {
		log.Printf("Error executing schema SQL: %v", err)
		return err
	}

	log.Println("Database schema created successfully.")

	// Run migrations
	err = runMigrations()
	if err != nil {
		log.Printf("Error running migrations: %v", err)
		return err
	}

	// Check if stocks table has data with retry
	var count int
	err = RetryQueryRow(DB, "SELECT COUNT(*) FROM stocks").Scan(&count)
	if err != nil {
		log.Printf("Error checking stocks count: %v", err)
		return err
	}

	// Only seed data if no stocks exist
	if count == 0 {
		log.Println("Seeding initial stock data...")
		_, err = RetryExec(DB, seedDataSQL)
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

// runMigrations runs database migrations
func runMigrations() error {
	log.Println("Running database migrations...")
	
	// Check if supabase_id column exists
	var columnExists bool
	err := RetryQueryRow(DB, `
		SELECT COUNT(*) > 0
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'users' 
		AND COLUMN_NAME = 'supabase_id'
	`).Scan(&columnExists)
	
	if err != nil {
		log.Printf("Error checking if supabase_id column exists: %v", err)
		return err
	}
	
	if !columnExists {
		log.Println("Adding supabase_id column to users table...")
		_, err = RetryExec(DB, "ALTER TABLE users ADD COLUMN supabase_id VARCHAR(255) NULL UNIQUE")
		if err != nil {
			log.Printf("Error adding supabase_id column: %v", err)
			return err
		}
		log.Println("Successfully added supabase_id column")
	} else {
		log.Println("supabase_id column already exists, skipping migration")
	}

	// Check if status column exists in stocks table
	var statusColumnExists bool
	err = RetryQueryRow(DB, `
		SELECT COUNT(*) > 0
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'stocks' 
		AND COLUMN_NAME = 'status'
	`).Scan(&statusColumnExists)
	
	if err != nil {
		log.Printf("Error checking if status column exists: %v", err)
		return err
	}
	
	if !statusColumnExists {
		log.Println("Adding status column to stocks table...")
		_, err = RetryExec(DB, "ALTER TABLE stocks ADD COLUMN status ENUM('active', 'distressed', 'delisted') DEFAULT 'active'")
		if err != nil {
			log.Printf("Error adding status column: %v", err)
			return err
		}
		log.Println("Successfully added status column")
	} else {
		log.Println("status column already exists, skipping migration")
	}

	// Check if sector_id column exists in stocks table
	var sectorColumnExists bool
	err = RetryQueryRow(DB, `
		SELECT COUNT(*) > 0
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'stocks' 
		AND COLUMN_NAME = 'sector_id'
	`).Scan(&sectorColumnExists)
	
	if err != nil {
		log.Printf("Error checking if sector_id column exists: %v", err)
		return err
	}
	
	if !sectorColumnExists {
		log.Println("Adding sector_id column to stocks table...")
		_, err = RetryExec(DB, "ALTER TABLE stocks ADD COLUMN sector_id INT NULL")
		if err != nil {
			log.Printf("Error adding sector_id column: %v", err)
			return err
		}

		// Add foreign key constraint
		_, err = RetryExec(DB, "ALTER TABLE stocks ADD FOREIGN KEY (sector_id) REFERENCES sectors(id)")
		if err != nil {
			log.Printf("Error adding sector_id foreign key: %v", err)
			return err
		}

		// Update existing stocks with sector IDs
		log.Println("Updating existing stocks with sector IDs...")
		sectorUpdates := []string{
			"UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Technology') WHERE symbol IN ('APPL', 'GOOG', 'AMZN', 'MSFT')",
			"UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Automotive') WHERE symbol = 'TSLA'",
			"UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Financial Services') WHERE symbol = 'JPM'",
			"UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Retail') WHERE symbol = 'WMT'",
			"UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Entertainment') WHERE symbol IN ('DIS', 'NFLX')",
			"UPDATE stocks SET sector_id = (SELECT id FROM sectors WHERE name = 'Healthcare') WHERE symbol = 'PFE'",
		}

		for _, update := range sectorUpdates {
			_, err = RetryExec(DB, update)
			if err != nil {
				log.Printf("Error updating stock sectors: %v", err)
				return err
			}
		}

		log.Println("Successfully added sector_id column and updated existing stocks")
	} else {
		log.Println("sector_id column already exists, skipping migration")
	}

	// Check if crisis_start column exists in stocks table
	var crisisStartExists bool
	err = RetryQueryRow(DB, `
		SELECT COUNT(*) > 0
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'stocks' 
		AND COLUMN_NAME = 'crisis_start'
	`).Scan(&crisisStartExists)
	
	if err != nil {
		log.Printf("Error checking if crisis_start column exists: %v", err)
		return err
	}
	
	if !crisisStartExists {
		log.Println("Adding crisis/bankruptcy columns to stocks table...")
		_, err = RetryExec(DB, `
			ALTER TABLE stocks 
			ADD COLUMN crisis_start TIMESTAMP NULL,
			ADD COLUMN recovery_chance FLOAT DEFAULT 0.03,
			ADD COLUMN bankruptcy_chance FLOAT DEFAULT 0.05
		`)
		if err != nil {
			log.Printf("Error adding crisis columns: %v", err)
			return err
		}
		log.Println("Successfully added crisis/bankruptcy columns")
	} else {
		log.Println("Crisis columns already exist, skipping migration")
	}

	// Check if username column exists in user_activity table
	var usernameColumnExists bool
	err = RetryQueryRow(DB, `
		SELECT COUNT(*) > 0
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'user_activity' 
		AND COLUMN_NAME = 'username'
	`).Scan(&usernameColumnExists)
	
	if err != nil {
		log.Printf("Error checking if username column exists in user_activity: %v", err)
		return err
	}
	
	if !usernameColumnExists {
		log.Println("Adding username column to user_activity table...")
		_, err = RetryExec(DB, "ALTER TABLE user_activity ADD COLUMN username VARCHAR(50) NOT NULL DEFAULT ''")
		if err != nil {
			log.Printf("Error adding username column to user_activity: %v", err)
			return err
		}

		// Update existing records with usernames from users table
		log.Println("Updating existing user_activity records with usernames...")
		_, err = RetryExec(DB, `
			UPDATE user_activity ua 
			JOIN users u ON ua.user_id = u.id 
			SET ua.username = u.username 
			WHERE ua.username = '' OR ua.username IS NULL
		`)
		if err != nil {
			log.Printf("Error updating existing user_activity records: %v", err)
			return err
		}

		// Add index for username column
		_, err = RetryExec(DB, "CREATE INDEX idx_user_activity_username ON user_activity(username)")
		if err != nil {
			log.Printf("Warning: Could not create index on username column: %v", err)
			// Don't return error as this is not critical
		}

		log.Println("Successfully added username column to user_activity table")
	} else {
		log.Println("username column already exists in user_activity, skipping migration")
	}

	// Check if monitoring columns exist in users table
	var monitoringColumnCount int
	err = RetryQueryRow(DB, `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'users' 
		AND COLUMN_NAME IN ('last_login', 'login_count', 'total_trades')
	`).Scan(&monitoringColumnCount)
	
	if err != nil {
		log.Printf("Error checking if monitoring columns exist in users table: %v", err)
		return err
	}
	
	if monitoringColumnCount < 3 {
		log.Println("Adding monitoring columns to users table...")
		
		// Check and add last_login column
		var lastLoginExists bool
		err = RetryQueryRow(DB, `
			SELECT COUNT(*) > 0
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'users' 
			AND COLUMN_NAME = 'last_login'
		`).Scan(&lastLoginExists)
		if err != nil {
			log.Printf("Error checking last_login column: %v", err)
			return err
		}
		if !lastLoginExists {
			_, err = RetryExec(DB, "ALTER TABLE users ADD COLUMN last_login TIMESTAMP NULL")
			if err != nil {
				log.Printf("Error adding last_login column: %v", err)
				return err
			}
		}

		// Check and add login_count column
		var loginCountExists bool
		err = RetryQueryRow(DB, `
			SELECT COUNT(*) > 0
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'users' 
			AND COLUMN_NAME = 'login_count'
		`).Scan(&loginCountExists)
		if err != nil {
			log.Printf("Error checking login_count column: %v", err)
			return err
		}
		if !loginCountExists {
			_, err = RetryExec(DB, "ALTER TABLE users ADD COLUMN login_count INT DEFAULT 0")
			if err != nil {
				log.Printf("Error adding login_count column: %v", err)
				return err
			}
		}

		// Check and add total_trades column
		var totalTradesExists bool
		err = RetryQueryRow(DB, `
			SELECT COUNT(*) > 0
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'users' 
			AND COLUMN_NAME = 'total_trades'
		`).Scan(&totalTradesExists)
		if err != nil {
			log.Printf("Error checking total_trades column: %v", err)
			return err
		}
		if !totalTradesExists {
			_, err = RetryExec(DB, "ALTER TABLE users ADD COLUMN total_trades INT DEFAULT 0")
			if err != nil {
				log.Printf("Error adding total_trades column: %v", err)
				return err
			}
		}

		log.Println("Successfully added monitoring columns to users table")
	} else {
		log.Println("All monitoring columns already exist in users table, skipping migration")
	}
	
	return nil
}
