package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Check if a token was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-jwt-secret.go <your-jwt-token>")
		os.Exit(1)
	}

	// Get the token from the command line arguments
	token := os.Args[1]

	// Split the token into its parts: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		fmt.Println("Invalid JWT token format. Expected 3 parts separated by dots.")
		os.Exit(1)
	}

	// Decode the header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		fmt.Printf("Error decoding header: %v\n", err)
		os.Exit(1)
	}

	// Parse the header
	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		fmt.Printf("Error parsing header: %v\n", err)
		os.Exit(1)
	}

	// Decode the payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Printf("Error decoding payload: %v\n", err)
		os.Exit(1)
	}

	// Parse the payload
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &header); err != nil {
		fmt.Printf("Error parsing payload: %v\n", err)
		os.Exit(1)
	}

	// Print token information
	fmt.Println("===== JWT Token Analysis =====")
	fmt.Println("Header:", string(headerBytes))
	fmt.Println("Payload:", string(payloadBytes))
	fmt.Printf("Signature (encoded): %s\n", parts[2])
	fmt.Println("\nSample JWT secrets that might work:")
	fmt.Println("1. your-secret-key-for-development-only")
	fmt.Println("2. OfficeStonksSecret")
	fmt.Println("3. stonkstoken")
	fmt.Println("4. yourjwtsecretkey")
	fmt.Println("5. admin123")
	fmt.Println("\nTo update the JWT secret in your code, modify the hardcoded value in internal/auth/jwt.go")
}