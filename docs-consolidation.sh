#!/bin/bash

# Documentation Consolidation Script for OfficeStonks
# This script consolidates documentation by category

# Set colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== OfficeStonks Documentation Consolidation ===${NC}"
echo -e "${YELLOW}This script will consolidate documentation files by category.${NC}"
read -p "Do you want to continue? (y/n): " -n 1 -r
echo # Move to a new line

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Consolidation aborted.${NC}"
    exit 1
fi

# Create docs directory if it doesn't exist
mkdir -p docs

# 1. Consolidate deployment documentation
echo -e "${BLUE}Consolidating deployment documentation...${NC}"
cat > docs/DEPLOYMENT.md << EOF
# Deployment Guide

This consolidated guide covers all aspects of deploying the OfficeStonks application.

## Table of Contents

1. [Railway Deployment](#railway-deployment)
2. [MySQL Configuration](#mysql-configuration)
3. [CORS and Proxy Configuration](#cors-and-proxy-configuration)
4. [Frontend Deployment](#frontend-deployment)
5. [Admin Dashboard Setup](#admin-dashboard-setup)

EOF

echo -e "${YELLOW}Adding railway deployment information...${NC}"
echo -e "## Railway Deployment\n" >> docs/DEPLOYMENT.md
cat RAILWAY_DEPLOYMENT.md | grep -v "^#" >> docs/DEPLOYMENT.md

echo -e "\n## MySQL Configuration\n" >> docs/DEPLOYMENT.md
cat RAILWAY_MYSQL_FIX.md | grep -v "^#" >> docs/DEPLOYMENT.md
cat MYSQL_DEPLOYMENT.md | grep -v "^#" >> docs/DEPLOYMENT.md

echo -e "\n## CORS and Proxy Configuration\n" >> docs/DEPLOYMENT.md
cat CORS_PROXY_INSTRUCTIONS.md | grep -v "^#" >> docs/DEPLOYMENT.md

echo -e "\n## Frontend Deployment\n" >> docs/DEPLOYMENT.md
cat FRONTEND_CORS_UPDATE.md | grep -v "^#" >> docs/DEPLOYMENT.md
cat FRONTEND_WEBSOCKET_CHANGES.md | grep -v "^#" >> docs/DEPLOYMENT.md

echo -e "\n## Admin Dashboard Setup\n" >> docs/DEPLOYMENT.md
cat ADMIN_JWT_FRONTEND_FIX.md | grep -v "^#" >> docs/DEPLOYMENT.md
cat ADMIN_API_FIX_DEPLOYMENT.md | grep -v "^#" >> docs/DEPLOYMENT.md

# 2. Consolidate project documentation
echo -e "${BLUE}Consolidating project documentation...${NC}"
cat > docs/PROJECT.md << EOF
# OfficeStonks Project Documentation

This document provides comprehensive information about the OfficeStonks project.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Project Structure](#project-structure)
3. [MVP Plan](#mvp-plan)
4. [MVP Accomplishments](#mvp-accomplishments)
5. [Next Steps](#next-steps)
6. [Testing](#testing)
7. [Contributing](#contributing)

EOF

echo -e "${YELLOW}Adding project overview...${NC}"
echo -e "## Project Overview\n" >> docs/PROJECT.md
cat Overview.md | grep -v "^#" >> docs/PROJECT.md

echo -e "\n## Project Structure\n" >> docs/PROJECT.md
cat PROJECT_STRUCTURE.md | grep -v "^#" >> docs/PROJECT.md

echo -e "\n## MVP Plan\n" >> docs/PROJECT.md
cat MVP_PLAN.md | grep -v "^#" >> docs/PROJECT.md
cat STEP_BY_STEP_MVP.md | grep -v "^#" >> docs/PROJECT.md

echo -e "\n## MVP Accomplishments\n" >> docs/PROJECT.md
cat MVP_ACCOMPLISHMENTS.md | grep -v "^#" >> docs/PROJECT.md

echo -e "\n## Next Steps\n" >> docs/PROJECT.md
cat MVP_NEXT_STEPS.md | grep -v "^#" >> docs/PROJECT.md

echo -e "\n## Testing\n" >> docs/PROJECT.md
cat TESTING.md | grep -v "^#" >> docs/PROJECT.md

echo -e "\n## Contributing\n" >> docs/PROJECT.md
cat CONTRIBUTING.md | grep -v "^#" >> docs/PROJECT.md

# 3. Consolidate cleanup documentation
echo -e "${BLUE}Consolidating cleanup documentation...${NC}"
cat > docs/CLEANUP.md << EOF
# Repository Cleanup

This document details the cleanup efforts for the OfficeStonks repository.

## Table of Contents

1. [Cleanup Summary](#cleanup-summary)
2. [Cleanup Process](#cleanup-process)
3. [Final State](#final-state)

EOF

echo -e "${YELLOW}Adding cleanup information...${NC}"
echo -e "## Cleanup Summary\n" >> docs/CLEANUP.md
cat CLEANUP_SUMMARY.md | grep -v "^#" >> docs/CLEANUP.md

echo -e "\n## Cleanup Process\n" >> docs/CLEANUP.md
cat REPOSITORY_CLEANUP.md | grep -v "^#" >> docs/CLEANUP.md 2>/dev/null
cat REPOSITORY_CLEANUP_UPDATE.md | grep -v "^#" >> docs/CLEANUP.md 2>/dev/null

echo -e "\n## Final State\n" >> docs/CLEANUP.md
cat REPOSITORY_CLEANUP_FINAL.md | grep -v "^#" >> docs/CLEANUP.md

# Create a manifest of consolidated files
echo -e "${BLUE}Creating documentation manifest...${NC}"
cat > docs/MANIFEST.md << EOF
# Documentation Manifest

This file lists all the consolidated documentation files and their sources.

## Consolidated Files

1. **DEPLOYMENT.md** - All deployment-related documentation
   - RAILWAY_DEPLOYMENT.md
   - RAILWAY_MYSQL_FIX.md
   - MYSQL_DEPLOYMENT.md
   - CORS_PROXY_INSTRUCTIONS.md
   - FRONTEND_CORS_UPDATE.md
   - FRONTEND_WEBSOCKET_CHANGES.md
   - ADMIN_JWT_FRONTEND_FIX.md
   - ADMIN_API_FIX_DEPLOYMENT.md

2. **PROJECT.md** - All project-related documentation
   - Overview.md
   - PROJECT_STRUCTURE.md
   - MVP_PLAN.md
   - STEP_BY_STEP_MVP.md
   - MVP_ACCOMPLISHMENTS.md
   - MVP_NEXT_STEPS.md
   - TESTING.md
   - CONTRIBUTING.md

3. **CLEANUP.md** - All cleanup-related documentation
   - CLEANUP_SUMMARY.md
   - REPOSITORY_CLEANUP.md
   - REPOSITORY_CLEANUP_UPDATE.md
   - REPOSITORY_CLEANUP_FINAL.md

## Original Files

The original documentation files have been preserved but are now considered deprecated.
Please refer to the consolidated documentation files above for the most up-to-date information.
EOF

echo -e "${GREEN}Documentation consolidation completed!${NC}"
echo -e "${BLUE}Consolidated files are available in the docs/ directory.${NC}"
echo -e "${YELLOW}Original files have been preserved for reference.${NC}"
echo -e "${YELLOW}To complete the consolidation, you may want to update README.md to point to the new documentation files.${NC}"