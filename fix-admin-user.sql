-- This script will directly add the is_admin column and create an admin user
-- Use this script to manually fix admin issues on Railway

-- First make sure the column exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;

-- Select user IDs to show existing users
SELECT id, username, is_admin FROM users;

-- Create admin user if it doesn't exist
INSERT INTO users (username, password_hash, cash_balance, is_admin)
SELECT 'admin', '$2a$10$JdvU7xXL6eLAC1ped9bY5.RMRxgNUT1Dg.Bh3ZJxXmVvIyAOKHYQu', 10000.00, TRUE
FROM dual
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');

-- Update an existing user to be admin - change 'admin' to an existing username if needed
UPDATE users SET is_admin = TRUE WHERE username = 'admin';

-- In case the admin user already exists but doesn't have admin flag
UPDATE users SET is_admin = TRUE WHERE username = 'admin';

-- Verify the admin user
SELECT id, username, is_admin FROM users WHERE username = 'admin';

-- Show all users with admin flag
SELECT id, username, is_admin FROM users WHERE is_admin = TRUE;