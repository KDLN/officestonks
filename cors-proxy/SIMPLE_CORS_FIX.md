# Simplified CORS Fix Deployment

This is an emergency fix for the CORS issues with the login and admin endpoints. The simplified version takes a more direct approach to setting CORS headers and handling preflight requests.

## Steps to Deploy the Simple Fix

1. Make sure the code is pulled from the repository (latest commit has the simple-cors-fix.js file)

2. Deploy to Railway with a modified start command:
   ```
   node simple-cors-fix.js
   ```

   You can do this by:
   - Going to the Railway Dashboard
   - Opening the CORS Proxy service
   - Going to Settings
   - Changing the "Start Command" to `node simple-cors-fix.js`
   - Clicking "Deploy"

3. Alternatively, use the Railway CLI:
   ```bash
   railway login
   cd /path/to/officestonks/cors-proxy
   railway up --service YOUR_SERVICE_NAME --start-command "node simple-cors-fix.js"
   ```

## How This Fix Works

The simplified fix:

1. Sets CORS headers directly on the response object for every request
2. Specifically handles OPTIONS preflight requests with a 204 status
3. Applies CORS headers even on proxy errors
4. Uses a more permissive approach to CORS that should work with any frontend
5. Allows the frontend URL: `https://officestonks-frontend-production.up.railway.app`

## Testing the Deployment

After deployment:

1. Test the health endpoint:
   ```
   curl https://officestonks-cors-proxy.up.railway.app/health
   ```

2. Test preflight for login:
   ```
   curl -X OPTIONS -i https://officestonks-cors-proxy.up.railway.app/api/auth/login -H "Origin: https://officestonks-frontend-production.up.railway.app"
   ```

3. Test preflight for admin routes:
   ```
   curl -X OPTIONS -i https://officestonks-cors-proxy.up.railway.app/api/admin/stocks/reset -H "Origin: https://officestonks-frontend-production.up.railway.app"
   ```

## Reverting if Needed

If the simple fix doesn't work, you can revert to the original proxy by changing the start command back to:
```
node index.js
```