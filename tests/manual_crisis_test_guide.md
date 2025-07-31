# Manual Crisis Testing Guide

## Prerequisites
1. Have at least 2-3 test user accounts with portfolios
2. Ensure users own shares in different stocks across sectors
3. Have admin access to the Admin Panel

## Step-by-Step Testing

### Test 1: Basic Crisis Event
1. **Login as Admin** and go to Admin Panel
2. **Navigate to Crisis Testing** section
3. **Select a stock** that users own (check Leaderboard for popular stocks)
4. **Click "Force Crisis Event"**
5. **Verify**:
   - Stock price drops to $0.01
   - Crisis news appears in News Feed
   - Stock shows as "distressed" in Stock List

### Test 2: Bankruptcy with Portfolio Impact
1. **Note current portfolios** of users who own the target stock
2. **Force Bankruptcy** on a stock with multiple holders
3. **Wait 5-10 seconds** for processing
4. **Verify**:
   - Stock disappears from all user portfolios
   - News feed shows bankruptcy announcement
   - Stock status shows as "delisted"
   - Users' total portfolio values decreased

### Test 3: Recovery Event
1. **Force Crisis** on a different stock
2. **Force Recovery** on the same stock
3. **Verify**:
   - Stock price jumps significantly (10x-100x)
   - Recovery news appears
   - Stock tradeable again at new price

### Test 4: Sector Contagion
1. **Note stocks** in the same sector (e.g., all Tech stocks)
2. **Force Bankruptcy** on one sector stock
3. **Monitor** other stocks in same sector
4. **Verify**:
   - Some sector peers show price drops
   - Sector contagion news may appear
   - Overall sector performance declines

### Test 5: Multiple User Impact
1. **Create scenario** where 3+ users own same stock
2. **Force Bankruptcy** on that stock
3. **Check each user's portfolio**
4. **Verify**:
   - All users lost their holdings
   - Portfolio values updated correctly
   - No orphaned holdings remain

## What to Look For

### ✅ Success Indicators:
- Smooth transitions between stock states
- Accurate portfolio updates
- Timely news generation
- No hanging or frozen states
- Consistent data across all views

### ❌ Potential Issues:
- Stock price not updating
- Holdings remaining after bankruptcy
- Missing news items
- Incorrect portfolio calculations
- UI not reflecting changes

## Post-Test Verification

1. **Check Database** (if you have access):
   ```sql
   -- Recent bankruptcies
   SELECT * FROM delisted_stocks ORDER BY delisted_at DESC LIMIT 5;
   
   -- Portfolio losses
   SELECT * FROM portfolio_losses ORDER BY delisted_at DESC LIMIT 10;
   
   -- Crisis news
   SELECT * FROM news_items WHERE type IN ('crisis', 'bankruptcy', 'recovery') 
   ORDER BY created_at DESC LIMIT 10;
   ```

2. **Check Logs** for any errors during processing

3. **Verify User Experience**:
   - Login as affected users
   - Confirm portfolio updates
   - Check transaction history
   - Ensure no phantom holdings

## Reporting Issues

If you find issues, note:
1. Exact steps to reproduce
2. Stock ID and symbol used
3. User accounts affected
4. Screenshots of unexpected behavior
5. Time of test (for log correlation)