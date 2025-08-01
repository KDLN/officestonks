# Monitoring System Setup Guide

## Overview
This guide helps you set up the comprehensive monitoring and user logging system for Office Stonks.

## Features Added

### Backend Features
- **User Session Tracking**: Tracks user logins, activity, and session duration
- **Activity Logging**: Detailed logging of all user actions (trades, logins, etc.)
- **System Metrics**: Real-time system performance monitoring
- **Request Tracking**: Automatic API request performance monitoring
- **WebSocket Connection Tracking**: Monitor active connections

### Frontend Features
- **Admin Monitoring Dashboard**: Comprehensive real-time dashboard at `/monitoring`
- **System Health Overview**: CPU, database, error rates, response times
- **Active Sessions View**: See who's online and their activity
- **User Activity Tracking**: Detailed activity logs with search and filtering
- **Real-time Updates**: Auto-refresh every 30 seconds

## Database Setup

### 1. Run the Monitoring Schema
Apply the new database schema for monitoring:

```bash
# Connect to your database and run the monitoring schema
mysql -u your_username -p your_database < monitoring_schema.sql
```

The schema adds these tables:
- `user_sessions` - Track user login sessions
- `user_activity` - Detailed activity logging
- `system_metrics` - Historical system performance data
- `rate_limit_violations` - Track abuse attempts
- `websocket_connections` - WebSocket connection tracking

### 2. Update Existing Tables
The script also adds columns to existing tables:
- `users` table: `last_login`, `login_count`, `total_trades`
- `transactions` table: `processing_time_ms`

## Backend Integration

### Files Added/Modified:
1. **New Models**: `internal/models/session.go`
2. **New Repositories**: 
   - `internal/repository/session_repository.go`
   - `internal/repository/activity_repository.go`
   - `internal/repository/metrics_repository.go`
3. **New Service**: `internal/services/monitoring_service.go`
4. **New Handler**: `internal/handlers/monitoring_handler.go`
5. **Updated**: `cmd/api/main.go` - Added monitoring routes and middleware

### API Endpoints Added:
- `GET /api/admin/monitoring/dashboard` - Complete dashboard data
- `GET /api/admin/monitoring/metrics` - System metrics only
- `GET /api/admin/monitoring/sessions` - Active sessions
- `GET /api/admin/monitoring/activity` - Recent activity
- `GET /api/admin/monitoring/user-activity?user_id=X` - User-specific activity
- `GET /api/admin/monitoring/user-sessions?user_id=X` - User sessions
- `GET /api/admin/monitoring/activity-range?start_time=X&end_time=Y` - Time range activity

## Frontend Integration

### Files Added:
1. **New Page**: `frontend/src/pages/MonitoringDashboard.js`
2. **New Styles**: `frontend/src/pages/MonitoringDashboard.css`
3. **Updated**: Navigation and App routing

### Access
- Navigate to `/monitoring` as an admin user
- New "📊 Monitoring" link in admin navigation

## Testing the System

### 1. Start the Application
```bash
# Backend
go run cmd/api/main.go

# Frontend (in separate terminal)
cd frontend && npm start
```

### 2. Test User Activity Tracking
1. Login as a regular user
2. Make some trades
3. Login as admin and visit `/monitoring`
4. Verify you can see:
   - Active sessions
   - Recent trade activity
   - System metrics

### 3. Test User Session Tracking
1. Login/logout multiple times
2. Check the monitoring dashboard for session data
3. Verify session duration and trade counts

### 4. Test System Health Monitoring
1. Monitor the dashboard for system health indicators
2. Check database connection status
3. Monitor error rates and response times

## Key Monitoring Features

### System Metrics Dashboard
- **Active Users**: Currently logged in users
- **Active Sessions**: Number of active sessions
- **Trades per Hour**: Trading volume
- **WebSocket Connections**: Real-time connections
- **Database Health**: Connection status
- **Error Rate**: Percentage of failed requests
- **Response Time**: Average API response time

### Active Sessions Table
- See all logged-in users
- View their IP addresses, login times, last activity
- Number of trades per session
- Quick access to user details

### Activity Feed
- Real-time feed of all user actions
- Success/failure indicators
- Error messages for failed actions
- IP address tracking
- Timestamp for all activities

### User Details Modal
- Click "View Details" on any user
- See their recent activities
- View their session history
- Monitor their trading patterns

## Performance Considerations

### Automatic Cleanup
- Expired sessions cleaned up every 10 minutes
- System metrics recorded every 5 minutes
- Activity logs automatically indexed for performance

### Rate Limiting
- Request tracking doesn't affect performance
- Metrics calculated in-memory for real-time data
- Database queries optimized with proper indexes

## Security Features

### Admin-Only Access
- All monitoring endpoints require admin privileges
- User activity tracking respects privacy (no sensitive data logged)
- IP address tracking for security purposes

### Abuse Detection
- Rate limit violations tracked
- Failed login attempts logged
- Suspicious activity patterns detectable

## Troubleshooting

### Database Connection Issues
Check the system health dashboard - database status will show "down" or "degraded" if there are connection issues.

### Missing Data
If monitoring data isn't appearing:
1. Verify database schema was applied correctly
2. Check server logs for errors
3. Ensure monitoring middleware is properly registered

### Performance Issues
The monitoring system is designed to be lightweight:
- Uses async logging to avoid blocking requests
- In-memory metrics for real-time data
- Efficient database queries with proper indexing

## Next Steps

### Potential Enhancements
1. **Alerts**: Email/Slack notifications for system issues
2. **Analytics**: Graphs and charts for historical data
3. **Export**: CSV export for activity logs
4. **Advanced Filtering**: More sophisticated search and filtering options
5. **Mobile Optimization**: Better mobile support for monitoring dashboard

The monitoring system provides comprehensive visibility into your Office Stonks application, helping you monitor user behavior, detect issues, and ensure optimal performance.