-- Safe Monitoring Schema for Office Stonks
-- This version handles cases where tables/columns might already exist

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
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_sessions_user_id (user_id),
    INDEX idx_user_sessions_active (is_active),
    INDEX idx_user_sessions_login_time (login_time)
);

-- Enhanced user activity tracking (more detailed than audit_logs)
CREATE TABLE IF NOT EXISTS user_activity (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
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
    INDEX idx_user_activity_action (action),
    INDEX idx_user_activity_timestamp (timestamp),
    INDEX idx_user_activity_success (success)
);

-- System performance metrics (for real-time monitoring)
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

-- Rate limiting violations (for abuse monitoring)
CREATE TABLE IF NOT EXISTS rate_limit_violations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL,
    user_id INT NULL,
    endpoint VARCHAR(255) NOT NULL,
    violation_count INT DEFAULT 1,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_rate_violations_ip (ip_address),
    INDEX idx_rate_violations_timestamp (timestamp)
);

-- WebSocket connection tracking
CREATE TABLE IF NOT EXISTS websocket_connections (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    session_id INT NULL,
    connection_id VARCHAR(255) NOT NULL,
    connected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    disconnected_at TIMESTAMP NULL,
    disconnect_reason VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES user_sessions(id) ON DELETE SET NULL,
    INDEX idx_websocket_user_id (user_id),
    INDEX idx_websocket_active (is_active)
);

-- Safely add columns to existing users table if they don't exist
-- Note: You may need to run these one at a time if your MySQL version doesn't support multiple ALTERs

-- Check and add last_login column
SET @dbname = DATABASE();
SET @tablename = 'users';
SET @columnname = 'last_login';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA=@dbname
      AND TABLE_NAME=@tablename
      AND COLUMN_NAME=@columnname
  ) > 0,
  'SELECT 1;',
  'ALTER TABLE users ADD COLUMN last_login TIMESTAMP NULL;'
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- Check and add login_count column
SET @columnname = 'login_count';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA=@dbname
      AND TABLE_NAME=@tablename
      AND COLUMN_NAME=@columnname
  ) > 0,
  'SELECT 1;',
  'ALTER TABLE users ADD COLUMN login_count INT DEFAULT 0;'
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- Check and add total_trades column
SET @columnname = 'total_trades';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA=@dbname
      AND TABLE_NAME=@tablename
      AND COLUMN_NAME=@columnname
  ) > 0,
  'SELECT 1;',
  'ALTER TABLE users ADD COLUMN total_trades INT DEFAULT 0;'
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- Safely add processing_time_ms to transactions table
SET @tablename = 'transactions';
SET @columnname = 'processing_time_ms';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA=@dbname
      AND TABLE_NAME=@tablename
      AND COLUMN_NAME=@columnname
  ) > 0,
  'SELECT 1;',
  'ALTER TABLE transactions ADD COLUMN processing_time_ms INT DEFAULT 0;'
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- Create views for common monitoring queries
CREATE OR REPLACE VIEW active_user_summary AS
SELECT 
    u.id,
    u.username,
    u.email,
    u.is_admin,
    u.cash_balance,
    u.last_login,
    s.ip_address,
    s.login_time,
    s.trades_count as session_trades,
    s.last_activity
FROM users u
JOIN user_sessions s ON u.id = s.user_id
WHERE s.is_active = TRUE
ORDER BY s.last_activity DESC;

CREATE OR REPLACE VIEW hourly_activity_summary AS
SELECT 
    DATE_FORMAT(timestamp, '%Y-%m-%d %H:00:00') as hour,
    COUNT(CASE WHEN action = 'login' THEN 1 END) as logins,
    COUNT(CASE WHEN action = 'trade' THEN 1 END) as trades,
    COUNT(CASE WHEN action = 'chat_message' THEN 1 END) as chat_messages,
    COUNT(CASE WHEN success = FALSE THEN 1 END) as errors,
    COUNT(*) as total_actions
FROM user_activity 
WHERE timestamp >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
GROUP BY DATE_FORMAT(timestamp, '%Y-%m-%d %H:00:00')
ORDER BY hour DESC;

-- Create indexes safely (will fail gracefully if they already exist)
-- We'll use a stored procedure to check if indexes exist first

DELIMITER $$

CREATE PROCEDURE create_index_if_not_exists(
    IN idx_name VARCHAR(255),
    IN tbl_name VARCHAR(255),
    IN col_name VARCHAR(255)
)
BEGIN
    DECLARE index_exists INT DEFAULT 0;
    
    SELECT COUNT(*) INTO index_exists
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = tbl_name
        AND INDEX_NAME = idx_name;
    
    IF index_exists = 0 THEN
        SET @sql = CONCAT('CREATE INDEX ', idx_name, ' ON ', tbl_name, '(', col_name, ')');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END$$

DELIMITER ;

-- Now create indexes using the procedure
CALL create_index_if_not_exists('idx_audit_logs_created_at', 'audit_logs', 'created_at');
CALL create_index_if_not_exists('idx_transactions_created_at', 'transactions', 'created_at');
CALL create_index_if_not_exists('idx_chat_messages_created_at', 'chat_messages', 'created_at');

-- Clean up the stored procedure
DROP PROCEDURE IF EXISTS create_index_if_not_exists;

-- Success message
SELECT 'Monitoring schema has been successfully applied!' AS message;