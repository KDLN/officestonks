-- Simple Monitoring Schema - Core Tables Only
-- Run this first to get monitoring working

-- User session tracking table
CREATE TABLE IF NOT EXISTS user_sessions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    login_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    logout_time TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    trades_count INT DEFAULT 0,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- User activity tracking
CREATE TABLE IF NOT EXISTS user_activity (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    username VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    details TEXT,
    ip_address VARCHAR(45) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT
);

-- System metrics
CREATE TABLE IF NOT EXISTS system_metrics (
    id INT AUTO_INCREMENT PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    active_users INT DEFAULT 0,
    total_sessions INT DEFAULT 0,
    active_sessions INT DEFAULT 0,
    trades_per_hour INT DEFAULT 0,
    websocket_connections INT DEFAULT 0,
    database_health VARCHAR(20) DEFAULT 'healthy',
    error_rate DECIMAL(5,4) DEFAULT 0.0000,
    avg_response_time_ms DECIMAL(8,2) DEFAULT 0.00
);

-- Add columns to users table if they don't exist
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS login_count INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_trades INT DEFAULT 0;

-- Basic indexes
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_activity_user_id ON user_activity(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_timestamp ON user_activity(timestamp);

SELECT 'Simple monitoring schema applied successfully!' AS result;