# Office Stonks Frontend Placeholder

This directory contains a simple placeholder for the Office Stonks frontend. Instead of trying to build the frontend during deployment, we're using a static placeholder page that links to the API.

## Deployment on Railway

1. Create a new service in your Railway project
2. Connect this directory to the service
3. Railway will deploy the static placeholder page

## Why This Approach?

Building the React frontend during deployment was causing issues because:

1. The frontend code was in a separate directory that wasn't accessible during the build
2. The build process required complex file copying between directories

This placeholder approach is a temporary solution until you can deploy the frontend properly.

## How to Deploy the Actual Frontend

For deploying the actual React frontend, you have several options:

### Option 1: Separate Repository

1. Move the frontend code to a separate repository
2. Deploy directly from that repository using Railway or Vercel
3. Set the `REACT_APP_API_URL` environment variable to point to your backend

### Option 2: Build Locally and Deploy Static Files

1. Build the frontend locally: `cd frontend && npm run build`
2. Deploy the contents of the `build` directory to a static hosting service

### Option 3: Use a Frontend-Specific Deployment Service

Services like Vercel, Netlify, or GitHub Pages are optimized for frontend deployment:

1. Connect your repository to one of these services
2. Configure the build command: `cd frontend && npm run build`
3. Set the output directory to `frontend/build`

## Backend API URL

The current placeholder links to:
`https://web-copy-production-5b48.up.railway.app/api`

Update this URL in `public/index.html` if your backend URL changes.