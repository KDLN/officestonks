-- Add is_admin column to users table
ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;

-- Create an admin user (username: admin, password: admin123)
-- Note: In production, use a strong password and ensure it's changed immediately
INSERT INTO users (username, password_hash, cash_balance, is_admin) 
VALUES ('admin', '$2a$10$JdvU7xXL6eLAC1ped9bY5.RMRxgNUT1Dg.Bh3ZJxXmVvIyAOKHYQu', 10000.00, TRUE);

-- Add admin endpoints permission function
-- This function will return true if the user is an admin
DELIMITER //
CREATE FUNCTION IF NOT EXISTS is_user_admin(user_id INT) RETURNS BOOLEAN
BEGIN
  DECLARE admin_status BOOLEAN;
  SELECT is_admin INTO admin_status FROM users WHERE id = user_id;
  RETURN admin_status;
END //
DELIMITER ;