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
   - Railway will detect the Dockerfile automatically

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

## Database Initialization

1. Connect to your Railway MySQL instance:
   - From your Railway dashboard, go to your MySQL service
   - Click "Connect" and use the provided connection details
   - You can use a MySQL client or the Railway CLI to run SQL commands

2. Run the schema initialization script:
   - Import the `backend/schema.sql` file to set up tables:
   ```bash
   # Using Railway CLI
   railway connect mysql
   # Then paste the contents of schema.sql

   # Or using a MySQL client
   mysql -h YOUR_MYSQL_HOST -u YOUR_MYSQL_USER -p YOUR_MYSQL_DATABASE < backend/schema.sql
   ```

## Troubleshooting Deployment Issues

### Common Issues and Solutions

1. **Database Connection Problems**:
   - Verify environment variables are correct
   - Check if the database user has proper permissions
   - Ensure your application is connecting to the right host and port

2. **Build Failures**:
   - Railway uses the Dockerfile for building
   - Check build logs for specific errors
   - Make sure all Go dependencies are properly specified in go.mod

3. **Runtime Errors**:
   - Check application logs in the Railway dashboard
   - Verify that environment variables are being correctly passed to your application
   - Make sure your application is listening on the correct port (PORT environment variable)

### Updating the Deployment

To update your deployment:

1. Make changes to your code
2. Commit and push to GitHub
3. Railway will automatically detect changes and start a new deployment

## Setting Up Custom Domain (Optional)

1. In Railway dashboard, go to your backend service
2. Navigate to "Settings" > "Domains"
3. Add your custom domain
4. Update your DNS settings as instructed by Railway

## Monitoring

1. Check deployment logs in Railway dashboard
2. Monitor your application's performance in the Railway metrics tab

## Next Steps

After successful deployment:

1. Access your API at the provided Railway URL or your custom domain
2. Set up the frontend (can be deployed separately on Railway or services like Vercel, Netlify)
3. Configure CORS settings to allow your frontend to communicate with the backend
4. Set up database backups