package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	fmt.Println("Admin Context Key Fix - Diagnostic Tool")
	fmt.Println("======================================")

	// Log the issue
	fmt.Println("The issue appears to be a mismatch in context keys between middleware and admin handler:")
	fmt.Println("- auth_middleware.go defines: const UserIDKey contextKey = \"userID\"")
	fmt.Println("- admin_handler.go uses: r.Context().Value(\"userID\").(int)")
	fmt.Println()
	fmt.Println("In Go, string keys and typed keys are not the same thing, even if the string value matches.")
	fmt.Println("We need to either:")
	fmt.Println("1. Use the correct contextKey type in admin_handler.go")
	fmt.Println("2. Export the contextKey from middleware to handlers package")
	fmt.Println("3. Change the middleware to use a string key")
	fmt.Println()

	// Display a sample JWT token from the request logs
	exampleToken := getExampleTokenFromArgs()
	if exampleToken != "" {
		fmt.Println("Example token found: ", truncateToken(exampleToken))
		fmt.Println("Attempting to parse token payload...")
		
		// Extract and parse the token payload
		parts := strings.Split(exampleToken, ".")
		if len(parts) != 3 {
			fmt.Println("ERROR: Invalid token format - expected 3 parts")
		} else {
			payload, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				fmt.Printf("ERROR: Failed to decode payload: %v\n", err)
			} else {
				fmt.Println("Payload (raw):", string(payload))
				
				// Pretty print the JSON
				var claims map[string]interface{}
				if err := json.Unmarshal(payload, &claims); err != nil {
					fmt.Printf("ERROR: Failed to parse JSON payload: %v\n", err)
				} else {
					jsonBytes, _ := json.MarshalIndent(claims, "", "  ")
					fmt.Println("Payload (formatted):")
					fmt.Println(string(jsonBytes))
					
					// Extract user_id for confirmation
					if userID, ok := claims["user_id"].(float64); ok {
						fmt.Printf("User ID: %.0f\n", userID)
					} else {
						fmt.Println("WARNING: user_id not found or not a number")
					}
				}
			}
		}
	}

	// Generate the code fix
	fmt.Println("\nProposed Fix:")
	fmt.Println("=============")
	fmt.Println("1. Export the context key from middleware package (in auth_middleware.go):")
	fmt.Println("```go")
	fmt.Println("// Key type for context values")
	fmt.Println("type ContextKey string")
	fmt.Println("")
	fmt.Println("// UserIDKey is the context key for the user ID - now exported")
	fmt.Println("const UserIDKey ContextKey = \"userID\"")
	fmt.Println("```")
	fmt.Println("")
	fmt.Println("2. Import middleware package in admin_handler.go and use the exported key:")
	fmt.Println("```go")
	fmt.Println("// Get user ID from context (set by auth middleware)")
	fmt.Println("userID, ok := r.Context().Value(middleware.UserIDKey).(int)")
	fmt.Println("```")
	fmt.Println("")
	fmt.Println("3. Alternative approach - modify admin_handler.go to use a string key (simpler fix):")
	fmt.Println("```go")
	fmt.Println("// Get user ID from context (set by auth middleware) - using string key")
	fmt.Println("userID, ok := r.Context().Value(contextKey(\"userID\")).(int)")
	fmt.Println("```")
	
	// Generate and save the fix scripts
	generateFixScriptFiles()
}

// Helper to truncate token for display
func truncateToken(token string) string {
	if len(token) > 20 {
		return token[:20] + "..."
	}
	return token
}

// Get example token from command line args if provided
func getExampleTokenFromArgs() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return ""
}

// Generate fix script files
func generateFixScriptFiles() {
	log.Println("Creating fix scripts...")
	
	// Create fix for auth_middleware.go
	exportScript := `#!/bin/bash
echo "Exporting UserIDKey from auth_middleware.go..."

# Replace the context key definitions
sed -i 's/type contextKey string/type ContextKey string/' /home/kdln/code/officestonks/internal/middleware/auth_middleware.go
sed -i 's/const UserIDKey contextKey = "userID"/const UserIDKey ContextKey = "userID"/' /home/kdln/code/officestonks/internal/middleware/auth_middleware.go

# Update references to contextKey within the file
sed -i 's/ctx := context.WithValue(r.Context(), UserIDKey, userID)/ctx := context.WithValue(r.Context(), UserIDKey, userID)/' /home/kdln/code/officestonks/internal/middleware/auth_middleware.go
sed -i 's/userID, ok := r.Context().Value(UserIDKey).(int)/userID, ok := r.Context().Value(UserIDKey).(int)/' /home/kdln/code/officestonks/internal/middleware/auth_middleware.go

echo "Middleware file updated to export UserIDKey"
`
	
	// Create fix for admin_handler.go
	handlerScript := `#!/bin/bash
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

echo "Admin handler file updated to use correct context key"
`

	// Create the combined fix script
	combinedScript := `#!/bin/bash
echo "Running combined fix for context key mismatch..."

# Export the context key from middleware
` + exportScript + `

# Update the admin handler to use the exported key
` + handlerScript + `

echo "Context key fix completed. Please rebuild and deploy the application."
`

	// Write the scripts to files
	os.WriteFile("fix-middleware-export.sh", []byte(exportScript), 0755)
	os.WriteFile("fix-admin-handler.sh", []byte(handlerScript), 0755)
	os.WriteFile("fix-context-keys.sh", []byte(combinedScript), 0755)
	
	fmt.Println("\nFix scripts created:")
	fmt.Println("- fix-middleware-export.sh: Updates middleware to export UserIDKey")
	fmt.Println("- fix-admin-handler.sh: Updates admin handler to use the exported key")
	fmt.Println("- fix-context-keys.sh: Runs both fixes together")
	fmt.Println("\nRun the combined fix with: ./fix-context-keys.sh")
}