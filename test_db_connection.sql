-- Test if basic tables exist
SHOW TABLES LIKE 'user_sessions';
SHOW TABLES LIKE 'user_activity';
SHOW TABLES LIKE 'users';

-- Test basic connectivity
SELECT 'Database connection is working!' as status, NOW() as timestamp;