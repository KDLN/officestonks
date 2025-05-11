#!/bin/bash

# Backup Cleanup Script for OfficeStonks
# This script cleans up the backup directories that are no longer needed

# Set colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== OfficeStonks Backup Cleanup ===${NC}"
echo -e "${YELLOW}This script will clean up the backup directories that are no longer needed.${NC}"
echo -e "${RED}WARNING: This is a destructive operation! Make sure you have a backup or commit your changes first.${NC}"
read -p "Do you want to continue? (y/n): " -n 1 -r
echo # Move to a new line

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Cleanup aborted.${NC}"
    exit 1
fi

# Record the initial repository size
INITIAL_SIZE=$(du -sh . | awk '{print $1}')
echo -e "${BLUE}Initial repository size: ${INITIAL_SIZE}${NC}"

echo -e "${BLUE}Creating archive of backup directories before removing them...${NC}"
mkdir -p archive
echo -e "${YELLOW}Archiving backup directory...${NC}"
tar -czf archive/backup-$(date +%Y%m%d).tar.gz backup
echo -e "${GREEN}Backup archived to archive/backup-$(date +%Y%m%d).tar.gz${NC}"

echo -e "${BLUE}Removing backup directories...${NC}"
echo -e "${YELLOW}Removing ./backup directory...${NC}"
rm -rf ./backup

# Check for other backup directories
echo -e "${YELLOW}Checking for other backup directories...${NC}"
find . -type d -name "*backup*" -not -path "*/node_modules/*" -not -path "*/archive/*" | while read dir; do
    echo -e "${YELLOW}Archiving $dir...${NC}"
    dirname=$(basename "$dir")
    tar -czf archive/$dirname-$(date +%Y%m%d).tar.gz "$dir"
    echo -e "${YELLOW}Removing $dir...${NC}"
    rm -rf "$dir"
done

# Record the final repository size
FINAL_SIZE=$(du -sh . | awk '{print $1}')
echo -e "${GREEN}Backup cleanup completed!${NC}"
echo -e "${BLUE}Initial size: ${INITIAL_SIZE}${NC}"
echo -e "${BLUE}Final size: ${FINAL_SIZE}${NC}"
echo -e "${YELLOW}Backups have been archived to the archive/ directory.${NC}"

echo -e "${BLUE}Creating a summary of the cleanup...${NC}"
cat > BACKUP_CLEANUP_SUMMARY.md << EOF
# Backup Cleanup Summary

## Overview

This document summarizes the backup cleanup performed on the OfficeStonks repository.

## Actions Performed

1. **Archived backup directories** - All backup directories were archived to the \`archive/\` directory before removal
2. **Removed backup directories** - The following directories were removed:
   - \`./backup\`
   - All directories with "backup" in their name

## Reason for Cleanup

The backup directories were taking up significant space in the repository and were no longer needed for active development. Archiving them preserves the content while reducing the repository size.

## Archive Location

All removed backups have been archived to the \`archive/\` directory in compressed format.

## Size Reduction

- Initial repository size: ${INITIAL_SIZE}
- Final repository size: ${FINAL_SIZE}

## Next Steps

If you need to access the archived backups, they can be extracted from the \`archive/\` directory.
EOF

echo -e "${GREEN}Backup cleanup summary written to BACKUP_CLEANUP_SUMMARY.md${NC}"