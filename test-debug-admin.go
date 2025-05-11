package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== Debug Admin API Test ===")

	// Base URL for the API
	baseURL := "https://web-production-1e26.up.railway.app"

	// Create a special debug token
	tokenPayload := fmt.Sprintf(`{"user_id":3,"exp":%d,"iat":%d,"debug_admin_access":true}`,
		time.Now().Add(24*time.Hour).Unix(),
		time.Now().Unix())
	
	// Encode in Base64 with invalid signature (JWT format)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + 
		encodeBase64(tokenPayload) + 
		".invalid_signature_that_will_be_bypassed"
	
	fmt.Println("Generated debug token:")
	fmt.Println(token)
	fmt.Println()

	// Test cases
	testCases := []struct {
		name    string
		path    string
		headers map[string]string
	}{
		{
			name:    "Query param: debug_admin_access=true",
			path:    "/api/admin/users?debug_admin_access=true&user_id=3",
			headers: map[string]string{},
		},
		{
			name:    "Query param: debug token",
			path:    "/api/admin/users?token=" + token,
			headers: map[string]string{},
		},
		{
			name: "Auth header: debug token",
			path: "/api/admin/users",
			headers: map[string]string{
				"Authorization": "Bearer " + token,
			},
		},
		{
			name: "Both debug token and debug_admin_access",
			path: "/api/admin/users?debug_admin_access=true&user_id=3",
			headers: map[string]string{
				"Authorization": "Bearer " + token,
			},
		},
	}

	// Run the tests
	client := &http.Client{Timeout: 10 * time.Second}
	
	for _, tc := range testCases {
		fmt.Printf("Testing: %s\n", tc.name)
		fmt.Printf("URL: %s%s\n", baseURL, tc.path)
		
		req, err := http.NewRequest("GET", baseURL+tc.path, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}
		
		// Add headers
		for k, v := range tc.headers {
			req.Header.Add(k, v)
			fmt.Printf("Header: %s: %s\n", k, v)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n\n", err)
			continue
		}
		
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Response: %s\n\n", string(body))
		
		// Pretty print JSON if possible
		var jsonData interface{}
		if json.Unmarshal(body, &jsonData) == nil {
			prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
			fmt.Printf("Formatted JSON:\n%s\n\n", string(prettyJSON))
		}
	}
}

func encodeBase64(input string) string {
	// Simple Base64 URL safe encoding
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var result strings.Builder

	for i := 0; i < len(input); i += 3 {
		var chunk uint32
		
		if i+2 < len(input) {
			chunk = (uint32(input[i]) << 16) | (uint32(input[i+1]) << 8) | uint32(input[i+2])
			result.WriteByte(base64Chars[(chunk>>18)&63])
			result.WriteByte(base64Chars[(chunk>>12)&63])
			result.WriteByte(base64Chars[(chunk>>6)&63])
			result.WriteByte(base64Chars[chunk&63])
		} else if i+1 < len(input) {
			chunk = (uint32(input[i]) << 16) | (uint32(input[i+1]) << 8)
			result.WriteByte(base64Chars[(chunk>>18)&63])
			result.WriteByte(base64Chars[(chunk>>12)&63])
			result.WriteByte(base64Chars[(chunk>>6)&63])
		} else {
			chunk = uint32(input[i]) << 16
			result.WriteByte(base64Chars[(chunk>>18)&63])
			result.WriteByte(base64Chars[(chunk>>12)&63])
		}
	}
	
	// No padding in URL-safe base64
	return result.String()
}