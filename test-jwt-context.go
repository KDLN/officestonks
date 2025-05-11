package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// Define both ways of storing context keys
type contextKey string
const TypedUserIDKey contextKey = "userID"
const StringUserIDKey = "userID"

func main() {
	fmt.Println("=== JWT Context Test ===")
	
	// Create a request
	req := httptest.NewRequest("GET", "/api/admin/users", nil)
	
	// Create contexts with different key types
	ctx1 := context.WithValue(req.Context(), TypedUserIDKey, 3)
	ctx2 := context.WithValue(req.Context(), StringUserIDKey, 3)
	ctx3 := context.WithValue(ctx1, StringUserIDKey, 3) // Both keys
	
	// Create requests with different contexts
	req1 := req.WithContext(ctx1)
	req2 := req.WithContext(ctx2)
	req3 := req.WithContext(ctx3)
	
	// Test retrieval with typed key
	fmt.Println("\nRetrieving with typed key (contextKey):")
	testRetrieval(req1, TypedUserIDKey, "req1") // Should work
	testRetrieval(req2, TypedUserIDKey, "req2") // Should fail
	testRetrieval(req3, TypedUserIDKey, "req3") // Should work
	
	// Test retrieval with string key
	fmt.Println("\nRetrieving with string key:")
	testRetrieval(req1, StringUserIDKey, "req1") // Should fail
	testRetrieval(req2, StringUserIDKey, "req2") // Should work
	testRetrieval(req3, StringUserIDKey, "req3") // Should work
	
	fmt.Println("\nThis demonstrates that context keys of different types,")
	fmt.Println("even with the same string value, are NOT interchangeable!")
	fmt.Println("The hotfix ensures both key types are used for compatibility.")
}

func testRetrieval(r *http.Request, key interface{}, reqName string) {
	if userID, ok := r.Context().Value(key).(int); ok {
		fmt.Printf("  ✅ %s: Found userID %d with key type %T\n", reqName, userID, key)
	} else {
		fmt.Printf("  ❌ %s: No userID found with key type %T\n", reqName, key)
	}
}
