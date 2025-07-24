# Office Stonks Playtest Guide 🎮

## 🎯 Overview
This guide provides comprehensive testing scenarios to validate all core features before MVP launch. Test both desktop and mobile experiences.

## 📋 Pre-Test Setup Checklist

### Environment Verification
- [ ] **Live Site Access**: Navigate to your Railway URL
- [ ] **Test Message Visible**: Look for "✨ New: Mobile responsive design + Light/Dark mode toggle! 🌓"
- [ ] **Theme Toggle Present**: Verify toggle switch in navigation bar
- [ ] **No Console Errors**: Open browser dev tools, check for JavaScript errors

---

## 🧪 Core Feature Testing

### 1. 🔐 Authentication System
**Objective**: Verify secure login/registration flow

#### Test Steps:
1. **Registration Flow**
   - [ ] Navigate to registration page
   - [ ] Create new account with unique email
   - [ ] Verify password requirements work
   - [ ] Confirm successful account creation
   - [ ] Auto-redirect to dashboard after registration

2. **Login Flow**
   - [ ] Log out if logged in
   - [ ] Log in with test credentials
   - [ ] Verify JWT token persistence (refresh page, stay logged in)
   - [ ] Test "remember me" functionality

3. **Security Testing**
   - [ ] Try accessing protected pages without login (should redirect to login)
   - [ ] Test logout functionality clears session
   - [ ] Verify admin pages require admin permissions

**Expected Results**: Secure authentication with proper redirects and session management.

---

### 2. 💰 Portfolio Management
**Objective**: Test portfolio tracking and real-time updates

#### Starting State Verification:
- [ ] **Initial Cash**: Verify $10,000 starting balance
- [ ] **Empty Portfolio**: No initial stock holdings
- [ ] **Zero Stock Value**: Portfolio shows $0 in stocks initially

#### Test Scenarios:

**A. Basic Portfolio Display**
- [ ] Total portfolio value shows correctly ($10,000 initially)
- [ ] Cash balance displays properly
- [ ] Stock value starts at $0
- [ ] Portfolio breakdown is accurate

**B. Real-time Updates**
- [ ] Portfolio values update without page refresh
- [ ] Stock holdings reflect recent trades
- [ ] Cash balance adjusts after trades
- [ ] Total value recalculates automatically

**Expected Results**: Accurate, real-time portfolio tracking with proper calculations.

---

### 3. 📈 Stock Trading System
**Objective**: Test complete buy/sell trading workflow

#### Test Portfolio:
Create a diverse test portfolio by making these trades:

**Trade Sequence 1: Basic Buying**
1. [ ] **Buy AAPL**: Purchase 10 shares of Apple
   - Verify total cost calculation
   - Confirm cash balance decreases
   - Check portfolio now shows AAPL holding

2. [ ] **Buy TSLA**: Purchase 5 shares of Tesla
   - Verify you can afford the purchase
   - Confirm portfolio shows multiple holdings
   - Check total portfolio value updates

3. [ ] **Buy MSFT**: Purchase 15 shares of Microsoft
   - Test buying with remaining cash
   - Verify portfolio diversification

**Trade Sequence 2: Selling**
4. [ ] **Sell AAPL**: Sell 5 shares (partial position)
   - Verify you can't sell more than you own
   - Confirm cash balance increases
   - Check remaining AAPL shares are correct

5. [ ] **Sell TSLA**: Sell all Tesla shares
   - Confirm position completely removed from portfolio
   - Verify cash balance reflects sale proceeds

**Trade Sequence 3: Edge Cases**
6. [ ] **Insufficient Funds**: Try to buy more than cash balance allows
   - Should show error message
   - Transaction should not execute

7. [ ] **Overselling**: Try to sell more shares than owned
   - Should show error message
   - Transaction should not execute

**Trade Confirmation Testing**
- [ ] **Confirmation Modal**: Verify modal appears for each trade
- [ ] **Trade Details**: Check stock, quantity, price, total are correct
- [ ] **Cancel Option**: Test canceling trades from confirmation modal
- [ ] **Confirm Execution**: Verify trades execute after confirmation

**Expected Results**: Smooth trading with proper validation and confirmation flow.

---

### 4. 🔄 Real-time Price Updates
**Objective**: Verify live market simulation and WebSocket connectivity

#### WebSocket Testing:
- [ ] **Connection Status**: Check browser dev tools for WebSocket connection
- [ ] **Price Updates**: Watch stock prices change every 2 seconds
- [ ] **Visual Indicators**: Verify price increase/decrease animations
- [ ] **Portfolio Impact**: Confirm portfolio value changes with price updates

#### Price Movement Validation:
- [ ] **Price Floor**: Verify no stock goes below $0.01
- [ ] **Realistic Changes**: Price movements seem reasonable (not extreme jumps)
- [ ] **Multiple Stocks**: Different stocks move independently
- [ ] **Impact Animation**: Price changes show green/red visual feedback

**Expected Results**: Stable WebSocket connection with realistic price movements and visual feedback.

---

### 5. 🏆 Social Features
**Objective**: Test leaderboard and chat functionality

#### Leaderboard Testing:
- [ ] **Ranking Display**: Leaderboard shows users by portfolio value
- [ ] **Real-time Updates**: Rankings update as portfolios change
- [ ] **User Position**: Find your position on leaderboard
- [ ] **Multiple Users**: Verify multiple users appear (if available)

#### Chat System:
- [ ] **Send Messages**: Send test messages in chat
- [ ] **Real-time Delivery**: Messages appear immediately for all users
- [ ] **Message History**: Previous messages load on page refresh
- [ ] **User Attribution**: Messages show correct sender names

**Expected Results**: Active social features enhancing competitive experience.

---

### 6. 🌓 Theme & Mobile Experience
**Objective**: Test new theme system and mobile responsiveness

#### Theme Testing:
- [ ] **Toggle Location**: Theme toggle visible in navigation
- [ ] **Light Mode**: Test light theme appearance
- [ ] **Dark Mode**: Switch to dark mode, verify all elements update
- [ ] **Persistence**: Refresh page, theme choice should persist
- [ ] **System Default**: Clear localStorage, should detect OS preference

#### Mobile Responsiveness Testing:
Use browser dev tools to test different screen sizes:

**Desktop (1200px+)**
- [ ] **Full Layout**: All elements visible and properly spaced
- [ ] **Navigation**: Horizontal nav bar with all links
- [ ] **Tables**: Full data tables display correctly
- [ ] **Trading**: Side-by-side trading interface

**Tablet (768px)**
- [ ] **Responsive Layout**: Elements stack appropriately
- [ ] **Navigation**: Nav elements wrap or stack vertically
- [ ] **Touch Targets**: Buttons are large enough for touch
- [ ] **Tables**: Tables remain usable or adapt layout

**Mobile (480px)**
- [ ] **Mobile Layout**: Full mobile optimization
- [ ] **Stacked Navigation**: Vertical navigation layout
- [ ] **Card Tables**: Tables convert to card layout
- [ ] **Trading Interface**: Stacked trading form sections
- [ ] **Touch Friendly**: All interactions work with touch

**Expected Results**: Seamless experience across all device sizes with proper theme switching.

---

### 7. 🔧 Admin Features (Admin Users Only)
**Objective**: Test admin panel functionality

#### Admin Access:
- [ ] **Admin Panel Link**: Verify admin link appears in navigation
- [ ] **User Management**: View all users list
- [ ] **Stock Reset**: Test stock price reset functionality
- [ ] **Chat Moderation**: Clear chat messages if needed

#### Emergency Features:
- [ ] **Stock Price Reset**: Reset all stocks to base prices
- [ ] **Market Control**: Verify atomic reset works without race conditions
- [ ] **User Management**: Edit user balances if needed

**Expected Results**: Admin tools work properly for system management.

---

## 🚨 Error Handling Testing

### Network Issues:
- [ ] **Offline Mode**: Disconnect internet, verify graceful degradation
- [ ] **Reconnection**: Reconnect internet, verify automatic recovery
- [ ] **Slow Connection**: Test on throttled connection

### Invalid Actions:
- [ ] **Invalid Stock ID**: Try accessing non-existent stock
- [ ] **Malformed Requests**: Verify API error handling
- [ ] **Session Expiry**: Test behavior when JWT expires

**Expected Results**: Graceful error handling with helpful user messages.

---

## 📊 Performance Testing

### Load Testing:
- [ ] **Multiple Tabs**: Open multiple tabs, verify performance
- [ ] **Rapid Trading**: Execute multiple trades quickly
- [ ] **WebSocket Stability**: Monitor connection under load
- [ ] **Memory Usage**: Check for memory leaks during extended use

### User Experience:
- [ ] **Page Load Speed**: Verify pages load under 3 seconds
- [ ] **Trade Execution**: Trades execute within 1 second
- [ ] **Real-time Updates**: Price updates don't lag
- [ ] **Smooth Animations**: Theme transitions and price changes are smooth

**Expected Results**: Responsive performance under normal usage patterns.

---

## 🎮 User Experience Scenarios

### New User Journey:
1. **First Visit**
   - [ ] Clear value proposition on landing
   - [ ] Easy registration process
   - [ ] Intuitive first trading experience

2. **Learning Curve**
   - [ ] UI is self-explanatory
   - [ ] Trading process is clear
   - [ ] Error messages are helpful

3. **Engagement**
   - [ ] Real-time updates create excitement
   - [ ] Leaderboard motivates competition
   - [ ] Chat enhances social experience

### Returning User Experience:
1. **Quick Re-engagement**
   - [ ] Fast login process
   - [ ] Portfolio immediately visible
   - [ ] Current market state clear

2. **Continued Play**
   - [ ] Easy to make quick trades
   - [ ] Portfolio tracking is satisfying
   - [ ] Social features encourage return visits

**Expected Results**: Engaging experience that encourages continued use.

---

## 🐛 Bug Report Template

When you find issues, document them using this format:

```
**Bug Title**: [Brief description]

**Steps to Reproduce**:
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Expected Result**: [What should happen]
**Actual Result**: [What actually happens]
**Browser/Device**: [Chrome/Firefox/Mobile/etc.]
**Theme**: [Light/Dark]
**Screenshots**: [If applicable]

**Severity**: [Critical/High/Medium/Low]
- Critical: Breaks core functionality
- High: Major feature doesn't work
- Medium: Minor feature issue
- Low: Cosmetic or edge case
```

---

## ✅ Pre-Launch Checklist

Before opening to beta testers, ensure:

### Core Functionality ✅
- [ ] All trading operations work correctly
- [ ] Portfolio tracking is accurate
- [ ] Real-time updates function properly
- [ ] Authentication is secure

### User Experience ✅
- [ ] Mobile responsive design works on all devices
- [ ] Theme switching works perfectly
- [ ] No critical bugs or console errors
- [ ] Performance is acceptable

### Social Features ✅
- [ ] Chat system works reliably
- [ ] Leaderboard updates correctly
- [ ] Multi-user experience is smooth

### Admin Tools ✅
- [ ] Admin panel accessible to admin users
- [ ] Stock reset functionality works
- [ ] User management tools function

---

## 🚀 Beta Testing Preparation

Once internal testing passes:

1. **Invite Limited Beta Users** (10-20 people)
2. **Monitor System Performance** during beta
3. **Collect User Feedback** systematically
4. **Document Common Issues** for fixes
5. **Iterate Based on Feedback** before wider launch

---

*This playtest guide ensures comprehensive validation of all Office Stonks features before launch. Happy testing! 🎯*