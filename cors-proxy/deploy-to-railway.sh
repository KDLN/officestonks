#!/bin/bash

# Script to deploy the CORS proxy to Railway

# Colors for better output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Deploying CORS Proxy to Railway${NC}"
echo -e "${YELLOW}================================${NC}"

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo -e "${RED}Railway CLI is not installed. Please install it first:${NC}"
    echo -e "npm i -g @railway/cli"
    exit 1
fi

# Check for Railway login
echo -e "${YELLOW}Checking Railway login status...${NC}"
railway whoami || {
    echo -e "${RED}You are not logged in to Railway. Please login first:${NC}"
    echo -e "railway login"
    exit 1
}

# Verify the current project
echo -e "${YELLOW}Verifying Railway project...${NC}"
railway project

# Deploy to Railway
echo -e "${YELLOW}Deploying CORS Proxy with enhanced admin CORS handling...${NC}"
echo -e "${YELLOW}This will deploy the directory: $(pwd)${NC}"
echo -e "Press CTRL+C to cancel or any key to continue..."
read -n 1 -s

# Deploy with Railway CLI
echo -e "${YELLOW}Starting deployment...${NC}"
railway up

# Deployment result
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Deployment completed successfully!${NC}"
    echo -e "${YELLOW}Important environment variables to check in Railway dashboard:${NC}"
    echo -e "- BACKEND_URL: Should point to your backend API service"
    echo -e "- JWT_SECRET: Should be set to a secure random string"
    echo -e "- PORT: Usually set automatically by Railway"
    echo -e "- Database variables if needed: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
    echo -e "- IPv4 variables: MYSQL_TCP_PROTOCOL=4, IPV6_DISABLED=true, GODEBUG=netdns=go"
    
    echo -e "\n${GREEN}Testing Instructions:${NC}"
    echo -e "1. Test a regular API endpoint: curl https://your-cors-proxy-url.up.railway.app/api/health"
    echo -e "2. Test an admin endpoint: curl -X OPTIONS -i https://your-cors-proxy-url.up.railway.app/api/admin/stocks/reset -H \"Origin: https://your-frontend-url.com\""
    echo -e "3. Check that preflight OPTIONS requests include proper CORS headers"
else
    echo -e "${RED}Deployment failed. Please check the errors above.${NC}"
fi