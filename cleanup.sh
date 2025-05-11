#!/bin/bash
# Repository cleanup script for OfficeStonks

set -e
echo "Starting OfficeStonks repository cleanup..."

# Create backup of current state
echo "Creating backup of key files..."
cp /home/kdln/code/officestonks/internal/repository/db.go /home/kdln/code/officestonks/internal/repository/db.go.backup
cp /home/kdln/code/officestonks/cmd/api/cors.go /home/kdln/code/officestonks/cmd/api/cors.go.backup

# 1. Replace core files with clean versions
echo "Replacing core files with clean versions..."
cp /home/kdln/code/officestonks/internal/repository/db.go.clean /home/kdln/code/officestonks/internal/repository/db.go
cp /home/kdln/code/officestonks/cmd/api/cors.go.clean /home/kdln/code/officestonks/cmd/api/cors.go

# 2. Remove temporary server implementations
echo "Removing temporary server implementations..."
rm -f /home/kdln/code/officestonks/server.fix.go
rm -f /home/kdln/code/officestonks/static-server.go
rm -f /home/kdln/code/officestonks/static-server
rm -f /home/kdln/code/officestonks/static-mode.go
rm -f /home/kdln/code/officestonks/build-static-server.sh
rm -f /home/kdln/code/officestonks/deploy-static-server.sh

# 3. Remove mock API files
echo "Removing mock API files..."
rm -rf /home/kdln/code/officestonks/mock-api
rm -f /home/kdln/code/officestonks/package-mock.json
rm -f /home/kdln/code/officestonks/cors-mock-server.js
rm -f /home/kdln/code/officestonks/deploy-mock-api.sh
rm -f /home/kdln/code/officestonks/deploy-ultra-simple.sh

# 4. Remove redundant Dockerfiles
echo "Removing redundant Dockerfiles..."
rm -f /home/kdln/code/officestonks/Dockerfile.static
rm -f /home/kdln/code/officestonks/Dockerfile.scratch
rm -f /home/kdln/code/officestonks/Dockerfile.simple
rm -f /home/kdln/code/officestonks/railway.static.json

# 5. Remove temporary CORS fix files
echo "Removing temporary CORS fix files..."
rm -f /home/kdln/code/officestonks/cors-proxy.js
rm -f /home/kdln/code/officestonks/cors-proxy-package.json
rm -f /home/kdln/code/officestonks/cors-proxy-setup.md
rm -f /home/kdln/code/officestonks/allow-all-proxy.js
rm -f /home/kdln/code/officestonks/allow-all-proxy-commonjs.js
rm -f /home/kdln/code/officestonks/allow-all-package.json
rm -f /home/kdln/code/officestonks/allow-all-package-commonjs.json
rm -f /home/kdln/code/officestonks/no-cors-proxy.js
rm -f /home/kdln/code/officestonks/no-cors-package.json
rm -f /home/kdln/code/officestonks/simple-cors-proxy.js
rm -f /home/kdln/code/officestonks/simple-package.json
rm -f /home/kdln/code/officestonks/ultra-simple-proxy.js
rm -f /home/kdln/code/officestonks/ultra-simple-package.json
rm -f /home/kdln/code/officestonks/cors-debug.html
rm -f /home/kdln/code/officestonks/cors-test.html
rm -f /home/kdln/code/officestonks/cors-verify.html
rm -f /home/kdln/code/officestonks/test-cors.sh
rm -f /home/kdln/code/officestonks/verify-cors.sh
rm -f /home/kdln/code/officestonks/test-simple-proxy.sh

# 6. Remove patches and frontend fixes
echo "Removing patches and frontend fixes..."
rm -f /home/kdln/code/officestonks/admin.js.patch
rm -f /home/kdln/code/officestonks/frontend-fix.js
rm -f /home/kdln/code/officestonks/quick-frontend-fix.js
rm -f /home/kdln/code/officestonks/final-frontend-fix.js
rm -f /home/kdln/code/officestonks/stock-reset-fix.js
rm -f /home/kdln/code/officestonks/test-mysql-connection.js

# 7. Remove redundant documentation
echo "Removing redundant documentation..."
rm -f /home/kdln/code/officestonks/ADMIN_FIX_SUMMARY.md
rm -f /home/kdln/code/officestonks/CORS-FIX-FOR-FRONTEND.md
rm -f /home/kdln/code/officestonks/CORS-FIX-README.md
rm -f /home/kdln/code/officestonks/CORS-PROXY-README.md
rm -f /home/kdln/code/officestonks/cors-allowed-backend.md

# Note about backup directory - leave this as a manual step
echo -e "\nNOTE: The backup directory has NOT been removed."
echo "After verifying that all necessary code has been migrated,"
echo "you can manually remove it with: rm -rf /home/kdln/code/officestonks/backup"

# Build the application to test
echo -e "\nBuilding the application to verify changes..."
go build -o server ./cmd/api/main.go

echo -e "\nCleanup complete! The repository has been cleaned and simplified."
echo "Next steps:"
echo "1. Test the application locally"
echo "2. Commit the changes"
echo "3. Deploy to Railway"
echo "4. Verify functionality in production"