-- Debug script to check monitoring system data

-- Check if tables exist
SHOW TABLES LIKE '%session%';
SHOW TABLES LIKE '%activity%';
SHOW TABLES LIKE '%metrics%';

-- Check user_sessions table
SELECT COUNT(*) as total_sessions FROM user_sessions;
SELECT * FROM user_sessions ORDER BY login_time DESC LIMIT 10;

-- Check active sessions
SELECT COUNT(*) as active_sessions FROM user_sessions WHERE is_active = TRUE;
SELECT u.username, s.* 
FROM user_sessions s 
JOIN users u ON s.user_id = u.id 
WHERE s.is_active = TRUE 
ORDER BY s.login_time DESC;

-- Check user_activity table
SELECT COUNT(*) as total_activities FROM user_activity;
SELECT * FROM user_activity ORDER BY timestamp DESC LIMIT 20;

-- Check system_metrics
SELECT * FROM system_metrics ORDER BY timestamp DESC LIMIT 5;

-- Check if users table has the new columns
DESCRIBE users;

-- Check recent logins from audit_logs
SELECT * FROM audit_logs WHERE action = 'login' ORDER BY created_at DESC LIMIT 10;

-- Check for any errors in user_activity
SELECT action, COUNT(*) as count, SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failures
FROM user_activity 
WHERE timestamp > DATE_SUB(NOW(), INTERVAL 1 HOUR)
GROUP BY action;

-- Debug: Show all tables in the database
SHOW TABLES;