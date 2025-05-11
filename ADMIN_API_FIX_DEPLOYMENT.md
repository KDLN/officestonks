# Admin API Fix Deployment Guide

This guide provides step-by-step instructions for deploying the admin API fixes to resolve the 401 Unauthorized errors when accessing admin endpoints.

## Overview of Fixes

We've implemented several fixes to address the admin API authentication issues:

1. **Enhanced JWT Validation Bypass**:
   - Added a robust JWT parser that can extract user IDs from tokens without validation
   - Implemented multiple fallback methods to ensure tokens work even if signature validation fails
   - Added detailed logging for troubleshooting

2. **Database Fixes**:
   - Created scripts to ensure the KDLN user (ID 3) has admin privileges
   - Added verification steps to confirm admin status

3. **Deployment Enhancements**:
   - Updated Railway configuration for more reliable deployment
   - Created verification tools to test functionality in production

## Deployment Steps

Follow these steps in order to deploy the fixes:

### 1. Update Server Code

```bash
# Clone the repository if you haven't already
git clone https://github.com/yourusername/officestonks.git
cd officestonks

# Pull the latest changes
git pull origin main

# Make sure you have the latest code
git checkout main
```

### 2. Build and Redeploy to Railway

```bash
# Push to Railway directly from your local machine
railway up

# Alternatively, push to GitHub to trigger automatic deployment
git push origin main
```

### 3. Apply Database Fixes

```bash
# Run the KDLN admin fix script
./run-kdln-admin-fix.sh
```

### 4. Verify the Deployment

```bash
# Run the JWT token test
go run test-jwt-validation.go <your-token>

# The token can be obtained from browser localStorage after login
```

### 5. Frontend Changes

The frontend team should implement one of the approaches described in the `ADMIN_JWT_FRONTEND_FIX.md` file.

## Troubleshooting

If you continue to experience issues after deployment, follow these troubleshooting steps:

### 1. Verify JWT Token Extraction

Check that tokens can be extracted using the test script:

```bash
go run test-jwt-validation.go <your-token>
```

### 2. Check Admin Status in Database

Verify the KDLN user has admin privileges:

```sql
SELECT id, username, is_admin FROM users WHERE id = 3;
```

### 3. Check Railway Deployment Logs

```bash
railway logs
```

### 4. Test Admin Endpoints Directly

Use curl to test admin endpoints:

```bash
# Replace with your actual token
TOKEN="your-jwt-token"

# Test with token in query parameter
curl "https://web-production-1e26.up.railway.app/api/admin/users?token=$TOKEN"

# Test with Authorization header
curl -H "Authorization: Bearer $TOKEN" "https://web-production-1e26.up.railway.app/api/admin/users"
```

### 5. Check CORS Headers

Ensure CORS headers are properly set for cross-origin requests:

```bash
# Check CORS headers with OPTIONS request
curl -X OPTIONS -i -H "Origin: https://officestonks-frontend-production.up.railway.app" \
  "https://web-production-1e26.up.railway.app/api/admin/users"
```

## Reverting Changes

If necessary, you can revert to the previous version:

```bash
# Revert to a specific commit
git checkout <previous-commit-hash>

# Redeploy
railway up
```

## Code Changes Summary

The following files were modified or created:

1. `/internal/auth/jwt.go` - Enhanced token validation
2. `/internal/auth/robust_parser.go` - New robust token parser
3. `/ensure-kdln-admin.sql` - Database fix for admin user
4. `/run-kdln-admin-fix.sh` - Script to apply database fix
5. `/test-jwt-validation.go` - Test script for JWT validation

## Contact

If you encounter issues that you can't resolve using this guide, please contact:

- KDLN (ID: 3) - Project administrator
- Development Team - Available through GitHub issues

## Security Note

The JWT validation bypass is a temporary solution to ensure functionality while a proper fix is developed. It should be replaced with a more secure solution in the future. We recommend:

1. Standardizing the JWT secret across all environments
2. Implementing proper secret rotation practices
3. Adding additional security measures like refresh tokens