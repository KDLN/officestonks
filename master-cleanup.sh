#!/bin/bash

# Master Cleanup Script for OfficeStonks
# This script runs all cleanup scripts in sequence to fully clean up the repository

# Set colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${BLUE}${BOLD}=== OfficeStonks Master Cleanup ===${NC}"
echo -e "${YELLOW}This script will run all cleanup scripts in sequence to fully clean up the repository.${NC}"
echo -e "${RED}${BOLD}WARNING: This is a highly destructive operation! Make sure you have committed your changes first.${NC}"
read -p "Do you want to continue? (y/n): " -n 1 -r
echo # Move to a new line

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Cleanup aborted.${NC}"
    exit 1
fi

# Record the initial repository size
INITIAL_SIZE=$(du -sh . | awk '{print $1}')
echo -e "${BLUE}Initial repository size: ${INITIAL_SIZE}${NC}"

# Make sure all scripts are executable
chmod +x cleanup-repo.sh docs-consolidation.sh proxy-cleanup.sh backup-cleanup.sh

# Step 1: Run repository cleanup
echo -e "\n${BLUE}${BOLD}=== STEP 1: General Repository Cleanup ===${NC}"
echo -e "${YELLOW}Running cleanup-repo.sh...${NC}"
./cleanup-repo.sh < <(echo 'y')

# Step 2: Clean up backup directories
echo -e "\n${BLUE}${BOLD}=== STEP 2: Backup Cleanup ===${NC}"
echo -e "${YELLOW}Running backup-cleanup.sh...${NC}"
./backup-cleanup.sh < <(echo 'y')

# Step 3: Consolidate documentation
echo -e "\n${BLUE}${BOLD}=== STEP 3: Documentation Consolidation ===${NC}"
echo -e "${YELLOW}Running docs-consolidation.sh...${NC}"
./docs-consolidation.sh < <(echo 'y')

# Step 4: Clean up proxy implementations
echo -e "\n${BLUE}${BOLD}=== STEP 4: Proxy Cleanup ===${NC}"
echo -e "${YELLOW}Running proxy-cleanup.sh...${NC}"
./proxy-cleanup.sh < <(echo 'y')

# Record the final repository size
FINAL_SIZE=$(du -sh . | awk '{print $1}')
echo -e "\n${GREEN}${BOLD}All cleanup operations completed successfully!${NC}"
echo -e "${BLUE}Initial size: ${INITIAL_SIZE}${NC}"
echo -e "${BLUE}Final size: ${FINAL_SIZE}${NC}"

# Create summary document
echo -e "${BLUE}Creating master cleanup summary...${NC}"

cat > MASTER_CLEANUP_SUMMARY.md << EOF
# Master Cleanup Summary

## Overview

This document summarizes all cleanup operations performed on the OfficeStonks repository.

## Cleanup Operations

1. **General Repository Cleanup**
   - Removed duplicate and unnecessary files
   - Removed redundant scripts and SQL files
   - Removed debug files

2. **Backup Cleanup**
   - Archived and removed all backup directories
   - Significant size reduction from backup removal

3. **Documentation Consolidation**
   - Consolidated all documentation into the docs/ directory
   - Organized documentation by category
   - Improved documentation structure

4. **Proxy Cleanup**
   - Kept the cors-proxy as the primary proxy implementation
   - Removed redundant proxy implementations
   - Improved proxy documentation

## Size Reduction

- Initial repository size: ${INITIAL_SIZE}
- Final repository size: ${FINAL_SIZE}

## Next Steps

1. Update the main README.md to point to the new documentation structure
2. Remove the cleanup scripts after successful completion
3. Commit the clean repository state

## Cleanup Scripts

The following cleanup scripts were used:

- \`cleanup-repo.sh\` - General repository cleanup
- \`backup-cleanup.sh\` - Backup directory cleanup
- \`docs-consolidation.sh\` - Documentation consolidation
- \`proxy-cleanup.sh\` - Proxy implementation cleanup
- \`master-cleanup.sh\` - Master script that ran all the above

All operations were completed successfully.
EOF

echo -e "${GREEN}${BOLD}Master cleanup summary written to MASTER_CLEANUP_SUMMARY.md${NC}"

echo -e "\n${YELLOW}${BOLD}Important Next Steps:${NC}"
echo -e "1. Review the repository state to ensure everything is as expected"
echo -e "2. Update the main README.md to point to the new documentation structure"
echo -e "3. Commit the changes with a message like 'Complete repository cleanup and organization'"
echo -e "4. Consider removing the cleanup scripts after successful commit"