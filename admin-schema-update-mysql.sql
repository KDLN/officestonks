-- Add is_admin column to users table if it doesn't exist
SET @s = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'users'
        AND COLUMN_NAME = 'is_admin'
    ),
    'SELECT "Column exists already"',
    'ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE'
));

PREPARE stmt FROM @s;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Update existing admin user if it exists or create one if it doesn't
-- The password hash corresponds to 'admin123' using bcrypt
INSERT INTO users (username, password_hash, cash_balance, is_admin)
SELECT 'admin', '$2a$10$JdvU7xXL6eLAC1ped9bY5.RMRxgNUT1Dg.Bh3ZJxXmVvIyAOKHYQu', 10000.00, TRUE
FROM dual
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');