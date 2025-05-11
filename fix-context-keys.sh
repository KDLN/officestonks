#!/bin/bash

echo "Running combined fix for context key mismatch..."

echo "Exporting UserIDKey from auth_middleware.go..."

# Replace the context key definitions to be exported
sed -i 's/type contextKey string/type ContextKey string/' /home/kdln/code/officestonks/internal/middleware/auth_middleware.go
sed -i 's/const UserIDKey contextKey = "userID"/const UserIDKey ContextKey = "userID"/' /home/kdln/code/officestonks/internal/middleware/auth_middleware.go

# Update the admin handler to use the correct key type
echo "Updating admin_handler.go to use the correct context key..."

# Import the middleware package if it's not already imported
grep -q "officestonks/internal/middleware" /home/kdln/code/officestonks/internal/handlers/admin_handler.go
if [ $? -ne 0 ]; then
  # Add the import
  sed -i '/import (/a\\t"officestonks/internal/middleware"' /home/kdln/code/officestonks/internal/handlers/admin_handler.go
  echo "Added middleware package import"
fi

# Replace the direct string key with the proper type
sed -i 's/userID, ok := r.Context().Value("userID").(int)/userID, ok := r.Context().Value(middleware.UserIDKey).(int)/' /home/kdln/code/officestonks/internal/handlers/admin_handler.go

# Also update other context value retrievals in the same file
sed -i 's/r.Context().Value("userID")/r.Context().Value(middleware.UserIDKey)/g' /home/kdln/code/officestonks/internal/handlers/admin_handler.go

# Also fix similar usages in other handler files
for file in /home/kdln/code/officestonks/internal/handlers/*.go; do
  if [ "$file" != "/home/kdln/code/officestonks/internal/handlers/admin_handler.go" ]; then
    # Add the import first if needed
    grep -q "officestonks/internal/middleware" "$file"
    if [ $? -ne 0 ]; then
      grep -q "import (" "$file"
      if [ $? -eq 0 ]; then
        sed -i '/import (/a\\t"officestonks/internal/middleware"' "$file"
      else
        # This file might have a single import, so need to convert it to multi-import format
        sed -i 's/import ".*"/import (\n\t"&"\n\t"officestonks\/internal\/middleware"\n)/' "$file"
      fi
    fi
    
    # Replace any usage of the string key
    sed -i 's/r.Context().Value("userID")/r.Context().Value(middleware.UserIDKey)/g' "$file"
  fi
done

# Make sure the GetUserID function also uses the right key type
sed -i 's/userID, ok := r.Context().Value(UserIDKey).(int)/userID, ok := r.Context().Value(UserIDKey).(int)/' /home/kdln/code/officestonks/internal/middleware/auth_middleware.go

echo "Context key fix completed. Please rebuild and deploy the application."