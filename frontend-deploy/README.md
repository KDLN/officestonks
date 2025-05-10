# Office Stonks Frontend Deployment

This directory contains the configuration for deploying the Office Stonks frontend as a static site.

## Deployment Options

### Option 1: Deploy on Railway

1. Create a new service in your Railway project
2. Connect this directory to the service
3. Railway will use the `railway.json` configuration to build and deploy the static site

### Option 2: Deploy on Vercel

1. Connect this repository to Vercel
2. Vercel will use the `vercel.json` configuration to build and deploy the static site

## Configuration

- `.env` - Contains environment variables for the frontend build, especially the API URL
- `package.json` - Contains scripts for building the frontend
- `railway.json` - Configuration for Railway deployment
- `vercel.json` - Configuration for Vercel deployment

## API URL

Make sure to update the `REACT_APP_API_URL` in the `.env` file to point to your deployed backend API URL.

## Local Testing

To test the frontend locally:

1. Clone the repository
2. Navigate to the frontend directory: `cd frontend`
3. Install dependencies: `npm install`
4. Start the development server: `npm start`

The frontend will connect to the API URL specified in `.env`.