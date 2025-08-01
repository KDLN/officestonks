-- Insert changelog entries for Office Stonks
-- Run this SQL in your Railway MySQL database

INSERT INTO changelog (version, title, description, changes, change_type, is_major, is_visible, created_at) VALUES
(
  'v1.0.0',
  'Office Stonks Launch',
  'Initial release of the multiplayer stock market simulation game.',
  '["Real-time stock trading with live price updates","Portfolio management and transaction history","Leaderboard rankings by portfolio value","Live chat system for social interaction","Admin controls for market management"]',
  'feature',
  true,
  true,
  '2024-12-18 00:00:00'
),
(
  'v1.1.0',
  'Market Sectors Foundation',
  'Introduced market sectors with correlated stock movements for more realistic trading.',
  '["Added 6 market sectors: Technology, Automotive, Financial Services, Retail, Entertainment, Healthcare","Stock prices now influenced by both individual trends (70%) and sector trends (30%)","Sector-wide correlations create realistic market behavior","Enhanced market simulator with sector tracking","Database schema updated to support sector relationships"]',
  'feature',
  true,
  true,
  '2024-12-25 00:00:00'
),
(
  'v1.2.0',
  'Crisis & News System',
  'Major update transforming crisis events into exciting high-stakes gameplay with comprehensive news coverage.',
  '["Price Zone Volatility: Penny stocks (10%), Low-cap (7%), Mid-cap (5%), Large-cap (3%) for realistic market behavior","Breaking News Ticker: Auto-rotating crisis alerts with play/pause controls on dashboard","Enhanced News Display: Filter by Crisis, Bankruptcy, Recovery, Sector with color-coded items and stock symbols","Portfolio Crisis Alerts: Real-time warnings for stocks at $0.01 with bankruptcy risk and recovery potential","Crisis Mechanics: 5% bankruptcy chance, 3% recovery chance every 2 seconds for $0.01 stocks","Trade Frequency Limiting: 5-second cooldown with 20 trades/hour limit per user for security","Database Integration: Sector foreign key relationships and complete schema for crisis tracking","Mobile Responsive: All new components optimized for mobile devices with smooth animations"]',
  'feature',
  true,
  true,
  '2025-01-01 00:00:00'
);

-- You can also add future entries as needed, for example:
-- INSERT INTO changelog (version, title, description, changes, change_type, is_major, is_visible, created_at) VALUES
-- (
--   'v1.2.1',
--   'Bug Fixes & Performance',
--   'Critical fixes for stability and performance improvements.',
--   '["Fixed TypeError in Tests.js formatDuration function","Improved changelog API fallback mechanism","Enhanced error handling in frontend components","Performance optimizations for real-time updates"]',
--   'bugfix',
--   false,
--   true,
--   NOW()
-- );