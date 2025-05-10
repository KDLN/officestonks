-- Update schema for existing database without dropping tables

-- Add is_admin column to users table if it doesn't exist
SET @s = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'users'
        AND COLUMN_NAME = 'is_admin'
    ),
    'SELECT "Column is_admin exists already"',
    'ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE'
));

PREPARE stmt FROM @s;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Insert admin user if it doesn't exist - password is 'admin123'
INSERT INTO users (username, password_hash, cash_balance, is_admin)
SELECT 'admin', '$2a$10$JdvU7xXL6eLAC1ped9bY5.RMRxgNUT1Dg.Bh3ZJxXmVvIyAOKHYQu', 10000.00, TRUE
FROM dual
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');

-- In case the admin user exists but doesn't have admin flag
UPDATE users SET is_admin = TRUE WHERE username = 'admin';

-- Verify the admin user exists with admin flag
SELECT id, username, is_admin FROM users WHERE username = 'admin';