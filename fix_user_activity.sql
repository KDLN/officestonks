-- Fix user_activity table to include username column
-- This will allow the monitoring endpoints to work properly

-- Add username column to user_activity table if it doesn't exist
ALTER TABLE user_activity 
ADD COLUMN IF NOT EXISTS username VARCHAR(50) NOT NULL DEFAULT '';

-- Update existing records with usernames from users table
UPDATE user_activity ua 
JOIN users u ON ua.user_id = u.id 
SET ua.username = u.username 
WHERE ua.username = '' OR ua.username IS NULL;

-- Add index for username if it doesn't exist
-- Note: This will fail gracefully if index already exists
CREATE INDEX idx_user_activity_username ON user_activity(username);