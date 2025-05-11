#!/bin/bash

# Repository Cleanup Script for OfficeStonks
# This script removes redundant, duplicate, and unnecessary files

# Set colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== OfficeStonks Repository Cleanup ===${NC}"
echo -e "${YELLOW}This script will clean up the repository by removing unnecessary files.${NC}"
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

# Create a list of files to be removed
echo -e "${BLUE}Creating list of files to be removed...${NC}"
rm -f cleanup-list.txt
touch cleanup-list.txt

# 1. Remove all backup directories
echo -e "${YELLOW}Identifying backup directories...${NC}"
find . -type d -name "backup" -not -path "*/node_modules/*" >> cleanup-list.txt

# 2. Remove duplicate database test files
echo -e "${YELLOW}Identifying duplicate database test files...${NC}"
find . -type f -name "test-db-*.go" -not -path "*/node_modules/*" -not -path "*/internal/*" | grep -v "test-db-resilience.go" >> cleanup-list.txt

# 3. Remove redundant proxy implementations
echo -e "${YELLOW}Identifying redundant proxy implementations...${NC}"
find ./minimal-proxy -type f >> cleanup-list.txt

# 4. Remove old documentation files
echo -e "${YELLOW}Identifying old documentation files...${NC}"
cat >> cleanup-list.txt << EOF
./REPOSITORY_CLEANUP.md
./REPOSITORY_CLEANUP_SUMMARY.md
./REPOSITORY_CLEANUP_UPDATE.md
EOF

# 5. Clean up temporary and debug files
echo -e "${YELLOW}Identifying temporary and debug files...${NC}"
find . -type f -name "debug-*.go" -not -path "*/node_modules/*" -not -path "*/cmd/*" >> cleanup-list.txt
find . -type f -name "*hardcoded*.go" -not -path "*/node_modules/*" >> cleanup-list.txt

# 6. Remove unused SQL files (keep only the main schema.sql and fix-admin-user.sql)
echo -e "${YELLOW}Identifying unused SQL files...${NC}"
find . -type f -name "*.sql" -not -path "*/node_modules/*" | grep -v "schema.sql" | grep -v "ensure-kdln-admin.sql" | grep -v "fix-admin-user.sql" >> cleanup-list.txt

# 7. Clean up redundant Dockerfiles
echo -e "${YELLOW}Identifying redundant Dockerfiles...${NC}"
ls Dockerfile.* | grep -v "Dockerfile.ipv4" >> cleanup-list.txt

# 8. Display the list of files to be removed
echo -e "${BLUE}The following files and directories will be removed:${NC}"
cat cleanup-list.txt | sort

# Confirmation
echo
echo -e "${RED}WARNING: This will delete all the files listed above!${NC}"
read -p "Are you sure you want to proceed? (y/n): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Cleanup aborted.${NC}"
    exit 1
fi

# Process the cleanup list
echo -e "${BLUE}Removing files...${NC}"
while IFS= read -r file; do
    if [ -e "$file" ]; then
        if [ -d "$file" ]; then
            echo -e "${YELLOW}Removing directory: ${file}${NC}"
            rm -rf "$file"
        else
            echo -e "${YELLOW}Removing file: ${file}${NC}"
            rm -f "$file"
        fi
    else
        echo -e "${RED}File not found: ${file}${NC}"
    fi
done < cleanup-list.txt

# Record the final repository size
FINAL_SIZE=$(du -sh . | awk '{print $1}')
echo -e "${GREEN}Repository cleanup completed!${NC}"
echo -e "${BLUE}Initial size: ${INITIAL_SIZE}${NC}"
echo -e "${BLUE}Final size: ${FINAL_SIZE}${NC}"

# Clean up the temporary file
rm -f cleanup-list.txt

# Create a record of the cleanup
cat > REPOSITORY_CLEANUP_FINAL.md << EOF
# Repository Cleanup Summary

## Removed Items

The following items were removed during the cleanup:

- Backup directories
- Duplicate database test files
- Redundant proxy implementations
- Old documentation files
- Temporary and debug files
- Unused SQL files
- Redundant Dockerfiles

## Reason for Cleanup

This cleanup was performed to:
1. Reduce repository size
2. Simplify the codebase
3. Remove redundant and duplicate files
4. Make the repository more maintainable

## Remaining Structure

The main codebase is now organized into the following directories:
- \`cmd/\`: Command-line applications
- \`internal/\`: Internal packages
- \`pkg/\`: Public packages
- \`frontend/\`: Frontend code
- \`cors-proxy/\`: CORS proxy for API requests

## Next Steps

Additional cleanup recommendations:
1. Consolidate the remaining documentation files
2. Remove unused dependencies
3. Refactor the code to use consistent patterns
4. Add more comprehensive tests
EOF

echo -e "${GREEN}Cleanup summary written to REPOSITORY_CLEANUP_FINAL.md${NC}"