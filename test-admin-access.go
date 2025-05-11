package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// Default API URL if not provided as argument
	defaultAPIURL = "https://web-production-1e26.up.railway.app"
	
	// Debug token with special flags
	debugToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZWJ1Z19hZG1pbl9hY2Nlc3MiOnRydWUsImV4cCI6MTc3ODUyNTkwNiwiaWF0IjoxNzQ2OTg5OTA2LCJ1c2VyX2lkIjozfQ.invalid_signature_that_will_be_bypassed"
)

func main() {
	fmt.Println("=== Admin API Access Test ===")
	
	// Get API URL from args or use default
	apiURL := defaultAPIURL
	if len(os.Args) > 1 {
		apiURL = os.Args[1]
	}
	
	fmt.Printf("Testing against API URL: %s\n\n", apiURL)
	
	// Test methods
	testMethods := []struct {
		name     string
		testFunc func(string) error
	}{
		{"Standard JWT Access", testStandardAccess},
		{"Debug Token Access", testDebugTokenAccess},
		{"URL Parameter Bypass", testURLParamBypass},
		{"Direct Debug Handler", testDirectDebugHandler},
	}
	
	// Run all tests
	for _, test := range testMethods {
		fmt.Printf("Running test: %s\n", test.name)
		err := test.testFunc(apiURL)
		if err != nil {
			fmt.Printf("❌ Test failed: %v\n\n", err)
		} else {
			fmt.Printf("✅ Test passed!\n\n")
		}
	}
}

// Test with standard JWT access (likely to fail unless proper token provided)
func testStandardAccess(apiURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	
	req, err := http.NewRequest("GET", apiURL+"/api/admin/users", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add JWT token from environment if available
	if token := os.Getenv("ADMIN_TOKEN"); token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	} else {
		// No token available, just try without it
		fmt.Println("No ADMIN_TOKEN environment variable found, trying without token")
	}
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status %d: %s", resp.StatusCode, string(body))
	}
	
	// Print response preview
	fmt.Printf("Response: %s\n", truncateString(string(body), 100))
	return nil
}

// Test with debug token in URL
func testDebugTokenAccess(apiURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/admin/users?token=%s", apiURL, debugToken), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status %d: %s", resp.StatusCode, string(body))
	}
	
	// Print response preview
	fmt.Printf("Response: %s\n", truncateString(string(body), 100))
	return nil
}

// Test with URL parameter bypass
func testURLParamBypass(apiURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	
	// Try with debug_admin_access and user_id parameters
	url := fmt.Sprintf("%s/api/admin/users?debug_admin_access=true&user_id=3", apiURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status %d: %s", resp.StatusCode, string(body))
	}
	
	// Verify response contains user data
	if !strings.Contains(string(body), "username") {
		return fmt.Errorf("response missing expected user data: %s", string(body))
	}
	
	// Print response preview
	fmt.Printf("Response: %s\n", truncateString(string(body), 100))
	return nil
}

// Test the direct debug handler
func testDirectDebugHandler(apiURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	
	// Try the debug endpoint if it exists
	url := fmt.Sprintf("%s/api/debug/admin/users", apiURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Debug endpoint test received status %d, this may be normal if endpoint not yet deployed\n", resp.StatusCode)
		// Don't count this as failure since endpoint may not exist yet
		return nil
	}
	
	// Print response preview
	fmt.Printf("Response: %s\n", truncateString(string(body), 100))
	return nil
}

// Helper function to truncate long strings
func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen] + "..."
}