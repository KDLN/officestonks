# Railway Deployment Instructions

## Railway Configuration

This project is set up to deploy on Railway.app. The following configurations have been made:

1. **Dockerfile**: The main Dockerfile builds the Go backend and sets up a proper execution environment.

2. **Start Script**: `/app/start-server.sh` is used to start the application.

3. **railway.json**: Contains the Railway-specific deployment configuration.

4. **Procfile**: Provides an alternative start command for Railway's Procfile-based deployments.

## Required Environment Variables

Make sure to set these environment variables in your Railway project:

```
DB_HOST=mysql.railway.internal
DB_USER=root
DB_PASSWORD=your-railway-provided-password
DB_NAME=railway
DB_PORT=3306
JWT_SECRET=your-jwt-secret-key
PORT=8080
```

## Database Setup

1. Add a MySQL service to your Railway project.
2. Railway will automatically populate most database environment variables.
3. Use the schema.sql file to initialize your database:
   - Connect to your Railway MySQL instance using the provided credentials
   - Import the schema.sql file

## Deployment Steps

1. Push your code to GitHub
2. Connect the repository to Railway
3. Railway will build and deploy the application automatically

## Troubleshooting

If your deployment fails:

1. Check the Railway logs for detailed error messages
2. Verify that all environment variables are set correctly
3. Ensure the database is properly initialized
4. Test your connection strings locally if possible

## Frontend Deployment

To deploy the frontend on Railway:

1. Add a separate service for the frontend
2. Set the environment variables for API connection
3. Configure the build and start commands for the frontend

## Manual Deployment

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login
railway login

# Link to your project
railway link

# Deploy
railway up
```

# CORS Proxy Deployment

This section covers how to deploy the simplified CORS proxy for admin API endpoints.

## Prerequisites

- Railway account
- Access to the officestonks repository

## Deployment Steps

1. Ensure you have the following files in your repository:
   - `simple-cors-proxy.js`: The main proxy server code
   - `simple-package.json`: Dependencies for the proxy
   - `Dockerfile.simple`: Docker configuration for the proxy

2. Create a new service in your Railway project:
   - Go to the Railway dashboard
   - Select "New Service" and choose "Deploy from Repo"
   - Select your repository

3. Configure the service:
   - Build Command: `cp simple-package.json package.json && npm install`
   - Start Command: `node simple-cors-proxy.js`

4. Set these environment variables:
   - `PORT`: 3001 (Railway will override this)
   - `API_BASE_URL`: `https://web-production-1e26.up.railway.app`

5. Click "Deploy"

## Testing the CORS Proxy

After deployment:

1. Use the `cors-debug.html` tool to verify the proxy is working
2. Test all endpoints, especially the admin endpoints
3. Check that OPTIONS preflight requests are handled correctly

## Troubleshooting

If you encounter CORS issues:

1. Check the Railway logs for detailed error information
2. Verify that the CORS headers are being set correctly in responses
3. Test OPTIONS requests directly to confirm proper preflight handling
4. Make sure the backend API URL is correct and accessible