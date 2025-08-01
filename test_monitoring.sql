-- Test monitoring system by manually inserting data
-- This will help us verify the tables work and identify issues

-- Test user_sessions table
INSERT INTO user_sessions (user_id, ip_address, user_agent) 
VALUES (1, '127.0.0.1', 'Test Browser');

-- Test user_activity table  
INSERT INTO user_activity (user_id, username, action, details, ip_address, success) 
VALUES (1, 'admin', 'test_login', 'Manual test entry', '127.0.0.1', true);

-- Verify data was inserted
SELECT 'user_sessions' as table_name, COUNT(*) as count FROM user_sessions
UNION ALL
SELECT 'user_activity' as table_name, COUNT(*) as count FROM user_activity;

-- Check the actual data
SELECT * FROM user_sessions;
SELECT * FROM user_activity;