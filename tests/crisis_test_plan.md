# Crisis/Recovery Mechanics Test Plan

## Test Scenarios

### 1. Single Stock Crisis Event
- **Test**: Force a stock to $0.01 and observe crisis mechanics
- **Expected**: 
  - Crisis news generated
  - 5% bankruptcy chance
  - 3% recovery chance
  - 92% stagnation

### 2. Bankruptcy with Portfolio Impact
- **Setup**: Create test users with holdings in target stock
- **Test**: Force bankruptcy on stock
- **Verify**:
  - Portfolio losses recorded
  - Stock removed from all portfolios
  - Stock marked as delisted
  - Bankruptcy news generated

### 3. Sector Contagion - Bankruptcy
- **Test**: Force bankruptcy in one sector stock
- **Verify**:
  - Other stocks in same sector show negative trend
  - Sector contagion news generated
  - 10% chance of crisis in vulnerable sector peers

### 4. Recovery Event
- **Test**: Force recovery on crisis stock
- **Expected**:
  - Stock price jumps 10x-100x
  - Recovery news generated
  - Positive sector contagion

### 5. Multi-Stock Sector Crisis
- **Test**: Force multiple stocks in same sector to crisis
- **Verify**:
  - Cascading effects
  - Multiple contagion events
  - Sector-wide news coverage

## Test Implementation Steps

1. **Create Test Data**
   - Add test users with varied portfolios
   - Ensure stocks across different sectors
   - Set up holdings in target stocks

2. **Execute Test Cases**
   - Use admin panel crisis testing buttons
   - Monitor logs for event processing
   - Check database tables for updates

3. **Verification Points**
   - `portfolio_losses` table entries
   - `delisted_stocks` table entries
   - `news_items` table for automated news
   - Portfolio updates for affected users
   - Stock status changes

## API Endpoints for Testing

- `POST /api/admin/crisis/force` - Force crisis event
- `POST /api/admin/crisis/bankruptcy` - Force bankruptcy
- `POST /api/admin/crisis/recovery` - Force recovery
- `GET /api/admin/crisis/status` - Get simulator status

## Database Queries for Verification

```sql
-- Check portfolio losses
SELECT * FROM portfolio_losses 
WHERE stock_id = ? 
ORDER BY delisted_at DESC;

-- Check delisted stocks
SELECT * FROM delisted_stocks 
WHERE reason = 'bankruptcy' 
ORDER BY delisted_at DESC;

-- Check news generation
SELECT * FROM news_items 
WHERE type IN ('crisis', 'bankruptcy', 'recovery') 
ORDER BY created_at DESC;

-- Check portfolio updates
SELECT * FROM portfolios 
WHERE stock_id = ?;
```