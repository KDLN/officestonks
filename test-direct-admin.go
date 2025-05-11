package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Claims represents the JWT token claims structure
type Claims struct {
	UserID int `json:"user_id"`
	Exp    int64 `json:"exp"`
	Iat    int64 `json:"iat"`
}

func main() {
	// Get token from command line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-direct-admin.go <jwt_token>")
		os.Exit(1)
	}
	
	token := os.Args[1]
	
	// Display token info
	fmt.Println("=== Token Information ===")
	fmt.Printf("Token (truncated): %s...\n", truncateToken(token))
	
	// Parse token payload
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		fmt.Println("Error: Invalid token format")
		os.Exit(1)
	}
	
	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Printf("Error decoding payload: %v\n", err)
		os.Exit(1)
	}
	
	// Parse claims
	var claims Claims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		fmt.Printf("Error parsing claims: %v\n", err)
		os.Exit(1)
	}
	
	// Display claims
	fmt.Println("\n=== Decoded Claims ===")
	fmt.Printf("User ID: %d\n", claims.UserID)
	fmt.Printf("Expires: %s\n", time.Unix(claims.Exp, 0).Format(time.RFC3339))
	fmt.Printf("Issued: %s\n", time.Unix(claims.Iat, 0).Format(time.RFC3339))
	
	// Check if token is expired
	if time.Now().Unix() > claims.Exp {
		fmt.Println("\n⚠️ WARNING: Token is expired!")
	} else {
		fmt.Println("\n✅ Token is not expired")
	}
	
	// Perform direct API tests
	fmt.Println("\n=== Testing Admin Endpoints ===")
	
	// Define base URL and endpoints to test
	baseURL := "https://web-production-1e26.up.railway.app"
	adminEndpoints := []string{
		"/api/admin/status",
		"/api/admin/users",
	}
	
	// Test each endpoint with different auth methods
	for _, endpoint := range adminEndpoints {
		fmt.Printf("\nTesting endpoint: %s\n", endpoint)
		
		// Method 1: Query parameter
		fmt.Println("  Method 1: Using token as query parameter...")
		url := fmt.Sprintf("%s%s?token=%s", baseURL, endpoint, token)
		response, statusCode, err := performRequest(url, "")
		printResult("Query parameter", statusCode, response, err)
		
		// Method 2: Authorization header
		fmt.Println("  Method 2: Using Authorization header...")
		url = fmt.Sprintf("%s%s", baseURL, endpoint)
		response, statusCode, err = performRequest(url, token)
		printResult("Auth header", statusCode, response, err)
		
		// Method 3: Both
		fmt.Println("  Method 3: Using both query parameter and header...")
		url = fmt.Sprintf("%s%s?token=%s", baseURL, endpoint, token)
		response, statusCode, err = performRequest(url, token)
		printResult("Both methods", statusCode, response, err)
	}
	
	// Add detailed debug information at the end
	fmt.Println("\n=== Additional Debug Information ===")
	fmt.Println("1. Check if admin privileges are properly set for user ID:", claims.UserID)
	fmt.Println("2. Verify the context key fix has been deployed to Railway")
	fmt.Println("3. Check if there are any CORS issues with OPTIONS preflight requests")
	fmt.Println("4. Examine server logs during these test requests")
}

// Helper function to make an HTTP request
func performRequest(url string, token string) (string, int, error) {
	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}
	
	// Add headers
	if token != "" && !strings.Contains(url, "token=") {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Origin", "https://officestonks-frontend-production.up.railway.app")
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}
	
	return string(body), resp.StatusCode, nil
}

// Helper to truncate token for display
func truncateToken(token string) string {
	if len(token) > 20 {
		return token[:20]
	}
	return token
}

// Helper to print result with color
func printResult(method string, statusCode int, response string, err error) {
	if err != nil {
		fmt.Printf("    ❌ %s: Error: %v\n", method, err)
		return
	}
	
	// Check if response was successful
	if statusCode >= 200 && statusCode < 300 {
		fmt.Printf("    ✅ %s: Status %d\n", method, statusCode)
		// Try to pretty print the JSON if possible
		var obj interface{}
		if err := json.Unmarshal([]byte(response), &obj); err == nil {
			jsonBytes, _ := json.MarshalIndent(obj, "        ", "  ")
			fmt.Printf("        %s\n", string(jsonBytes))
		} else {
			fmt.Printf("        %s\n", response)
		}
	} else {
		fmt.Printf("    ❌ %s: Status %d\n", method, statusCode)
		fmt.Printf("        %s\n", response)
	}
}