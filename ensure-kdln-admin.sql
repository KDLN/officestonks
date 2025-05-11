-- Script to ensure KDLN user (ID 3) has admin privileges
-- This is a targeted fix for the current admin access issue

-- First, show user information for debugging
SELECT 'Current KDLN user status:' AS message;
SELECT id, username, is_admin, cash_balance FROM users WHERE id = 3;

-- Make sure is_admin column exists (should already exist, but just to be safe)
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;

-- Update KDLN user to have admin privileges
UPDATE users SET is_admin = TRUE WHERE id = 3;

-- If user doesn't exist, create it (should never happen, but as a fallback)
INSERT INTO users (id, username, password_hash, cash_balance, is_admin)
SELECT 3, 'KDLN', '$2a$10$JdvU7xXL6eLAC1ped9bY5.RMRxgNUT1Dg.Bh3ZJxXmVvIyAOKHYQu', 10000.00, TRUE
FROM dual
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 3);

-- Show updated user information to confirm changes
SELECT 'Updated KDLN user status:' AS message;
SELECT id, username, is_admin, cash_balance FROM users WHERE id = 3;

-- Check if there are any users with admin privileges
SELECT 'All admin users:' AS message;
SELECT id, username, is_admin FROM users WHERE is_admin = TRUE;