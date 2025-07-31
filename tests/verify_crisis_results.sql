-- Crisis Mechanics Verification Queries
-- Run these after testing to verify the system worked correctly

-- 1. Check recent bankruptcies
SELECT 'Recent Bankruptcies' as query_type;
SELECT 
    id,
    symbol,
    name,
    sector,
    final_price,
    delisted_at,
    reason
FROM delisted_stocks
WHERE reason = 'bankruptcy'
ORDER BY delisted_at DESC
LIMIT 10;

-- 2. Check portfolio losses by stock
SELECT 'Portfolio Losses by Stock' as query_type;
SELECT 
    stock_symbol,
    stock_name,
    COUNT(DISTINCT user_id) as affected_users,
    SUM(quantity) as total_shares_lost,
    SUM(loss_amount) as total_loss_amount,
    MAX(delisted_at) as bankruptcy_date
FROM portfolio_losses
GROUP BY stock_id, stock_symbol, stock_name
ORDER BY bankruptcy_date DESC;

-- 3. Check user losses
SELECT 'Top User Losses' as query_type;
SELECT 
    pl.user_id,
    u.username,
    COUNT(DISTINCT pl.stock_id) as stocks_lost,
    SUM(pl.loss_amount) as total_losses
FROM portfolio_losses pl
JOIN users u ON pl.user_id = u.id
GROUP BY pl.user_id, u.username
ORDER BY total_losses DESC
LIMIT 10;

-- 4. Check crisis/bankruptcy news
SELECT 'Crisis Related News' as query_type;
SELECT 
    id,
    type,
    title,
    stock_id,
    created_at
FROM news_items
WHERE type IN ('crisis', 'bankruptcy', 'recovery')
ORDER BY created_at DESC
LIMIT 20;

-- 5. Check stocks in crisis (at $0.01)
SELECT 'Stocks Currently in Crisis' as query_type;
SELECT 
    id,
    symbol,
    name,
    sector,
    current_price,
    status,
    crisis_start,
    recovery_chance,
    bankruptcy_chance
FROM stocks
WHERE current_price <= 0.01 
  AND status = 'distressed'
ORDER BY crisis_start DESC;

-- 6. Sector impact analysis
SELECT 'Sector Impact Analysis' as query_type;
SELECT 
    s.sector,
    COUNT(DISTINCT CASE WHEN s.status = 'distressed' THEN s.id END) as distressed_stocks,
    COUNT(DISTINCT CASE WHEN s.status = 'delisted' THEN s.id END) as delisted_stocks,
    COUNT(DISTINCT CASE WHEN s.current_price <= 0.01 THEN s.id END) as crisis_stocks,
    AVG(s.current_price) as avg_price,
    MIN(s.current_price) as min_price
FROM stocks s
GROUP BY s.sector
ORDER BY distressed_stocks DESC;

-- 7. Recent portfolio changes
SELECT 'Portfolios Affected by Bankruptcies' as query_type;
SELECT 
    p.user_id,
    u.username,
    COUNT(DISTINCT p.stock_id) as current_holdings,
    u.cash_balance
FROM portfolios p
JOIN users u ON p.user_id = u.id
WHERE p.user_id IN (
    SELECT DISTINCT user_id 
    FROM portfolio_losses 
    WHERE delisted_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)
)
GROUP BY p.user_id, u.username, u.cash_balance;

-- 8. Verify no bankrupt stocks remain in portfolios
SELECT 'Bankrupt Stocks Still in Portfolios (Should be 0)' as query_type;
SELECT COUNT(*) as invalid_holdings
FROM portfolios p
JOIN stocks s ON p.stock_id = s.id
WHERE s.status = 'delisted';