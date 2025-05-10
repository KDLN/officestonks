#!/bin/bash

echo "=== REPOSITORY STRUCTURE VERIFICATION ==="
echo "Checking that all necessary files are in the right place..."

# Check Go files in root structure
if [ -f /home/kdln/code/officestonks/cmd/api/main.go ]; then
  echo "✅ main.go exists in cmd/api/"
else
  echo "❌ main.go not found in cmd/api/"
  exit 1
fi

# Check internal directory structure
echo "Checking internal directory structure..."
REQUIRED_DIRS=(
  "/home/kdln/code/officestonks/internal/auth"
  "/home/kdln/code/officestonks/internal/handlers"
  "/home/kdln/code/officestonks/internal/middleware"
  "/home/kdln/code/officestonks/internal/models"
  "/home/kdln/code/officestonks/internal/repository"
  "/home/kdln/code/officestonks/internal/services"
  "/home/kdln/code/officestonks/internal/websocket"
)

for dir in "${REQUIRED_DIRS[@]}"; do
  if [ -d "$dir" ]; then
    echo "✅ Directory $dir exists"
  else
    echo "❌ Directory $dir is missing"
    exit 1
  fi
done

# Check key files
echo "Checking key files..."
REQUIRED_FILES=(
  "/home/kdln/code/officestonks/schema.sql"
  "/home/kdln/code/officestonks/go.mod"
  "/home/kdln/code/officestonks/go.sum"
  "/home/kdln/code/officestonks/start-server.sh"
  "/home/kdln/code/officestonks/init-db.sh"
  "/home/kdln/code/officestonks/start.sh"
  "/home/kdln/code/officestonks/Dockerfile"
)

for file in "${REQUIRED_FILES[@]}"; do
  if [ -f "$file" ]; then
    echo "✅ File $file exists"
  else
    echo "❌ File $file is missing"
    exit 1
  fi
done

# Check permissions on scripts
echo "Checking script permissions..."
SCRIPT_FILES=(
  "/home/kdln/code/officestonks/start-server.sh"
  "/home/kdln/code/officestonks/init-db.sh"
  "/home/kdln/code/officestonks/start.sh"
)

for script in "${SCRIPT_FILES[@]}"; do
  if [ -x "$script" ]; then
    echo "✅ Script $script is executable"
  else
    echo "❌ Script $script is not executable"
    chmod +x "$script"
    echo "   → Made script executable"
  fi
done

# Verify middleware files
if [ -f /home/kdln/code/officestonks/internal/middleware/rate_limiter.go ]; then
  echo "✅ Rate limiter middleware is present"
else
  echo "❌ Rate limiter middleware is missing"
fi

# Attempt to build the application
echo "Attempting to build the application..."
cd /home/kdln/code/officestonks
if go build -v ./cmd/api/main.go; then
  echo "✅ Application builds successfully"
  # Clean up the binary
  rm -f main
else
  echo "❌ Application build failed"
  exit 1
fi

# Check for duplicate code in backend/backend
if [ -d "/home/kdln/code/officestonks/backend/backend" ]; then
  echo "⚠️ Warning: Nested backend directory still exists and may contain duplicate code"
  echo "   Consider removing it after verifying all code is migrated properly"
fi

echo "✅ Repository structure verification complete!"
echo "The repository is properly structured for deployment."