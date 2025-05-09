# Deployment Guide for Office Stonks

This guide explains how to deploy Office Stonks using GitHub and Railway.

## Prerequisites

1. [GitHub account](https://github.com)
2. [Railway account](https://railway.app)

## Setup GitHub Repository

1. Create a new GitHub repository
   - Go to https://github.com/new
   - Name: `officestonks`
   - Initialize with a README

2. Push existing code to the repository:
   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   git branch -M main
   git remote add origin https://github.com/yourusername/officestonks.git
   git push -u origin main
   ```

## Setup Railway Project

1. Create a new project in Railway:
   - Visit [Railway Dashboard](https://railway.app/dashboard)
   - Click "New Project"
   - Select "Deploy from GitHub repo"

2. Connect your GitHub repository:
   - Select the `officestonks` repository
   - Railway will automatically detect the configuration

3. Set up MySQL service:
   - In your project, click "New Service" > "Database" > "MySQL"
   - Railway will provide connection details in the Variables tab

4. Configure environment variables:
   - Go to your backend service's "Variables" tab
   - Add the following variables (use MySQL service's values):
     ```
     DB_HOST=
     DB_USER=
     DB_PASSWORD=
     DB_NAME=
     DB_PORT=3306
     JWT_SECRET=your-secret-key
     PORT=8080
     ```

5. Setup GitHub Secrets for CI/CD:
   - Go to your GitHub repository
   - Navigate to Settings > Secrets > Actions
   - Add a new secret: `RAILWAY_TOKEN`
   - Get your token from Railway CLI by running:
     ```bash
     railway login
     railway whoami
     ```

## Database Initialization

1. Connect to your Railway MySQL instance:
   - Get connection details from Railway dashboard
   - Use a MySQL client to connect

2. Run the schema initialization script:
   - Import the `schema.sql` file to set up tables

## Frontend Deployment

For the MVP, we'll use a separate frontend deployment on Railway:

1. Create a new service in your Railway project:
   - Click "New Service" > "Empty Service"
   - Set the build command to: `cd frontend && npm install && npm run build`
   - Set the start command to: `cd frontend && npx serve -s build`

2. Configure environment variables:
   - Add `REACT_APP_API_URL` pointing to your backend service URL
   - You can find this URL in the Railway dashboard

## Continuous Integration/Deployment

Our GitHub Actions workflows handle CI/CD:

1. **Test Workflow** (`test.yml`):
   - Runs on every push and pull request
   - Tests both backend and frontend code
   - Ensures everything works before deployment

2. **Deploy Workflow** (`deploy.yml`):
   - Runs on pushes to main branch
   - Automatically deploys to Railway

## Manual Deployment

If you need to deploy manually, use Railway CLI:

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

## Monitoring

1. Check deployment logs in Railway dashboard
2. Monitor your application's performance in the Railway metrics tab

## Troubleshooting

1. **Database Connection Issues**:
   - Verify environment variables are correct
   - Check firewall settings in Railway

2. **Failed Deployments**:
   - Check GitHub Actions logs for errors
   - Verify Railway configurations

3. **Application Errors**:
   - Check Railway logs for backend/frontend services

## Next Steps

After initial deployment:

1. Set up a custom domain (in Railway settings)
2. Configure SSL/TLS (automatic with Railway)
3. Set up monitoring and alerts
4. Implement database backups