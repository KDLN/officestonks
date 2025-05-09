package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yourusername/officestonks/internal/models"
)

func TestRegisterUser(t *testing.T) {
	// Skip if no test database connection
	if TestDB == nil {
		t.Skip("No test database connection")
	}

	// Setup test router
	router := SetupTestRouter(TestDB)

	// Test cases
	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		shouldSucceed  bool
	}{
		{"ValidRegister", "testuser1", "password123", http.StatusCreated, true},
		{"DuplicateUsername", "testuser1", "password123", http.StatusBadRequest, false},
		{"EmptyUsername", "", "password123", http.StatusBadRequest, false},
		{"EmptyPassword", "testuser2", "", http.StatusBadRequest, false},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request body
			reqBody := models.AuthRequest{
				Username: tc.username,
				Password: tc.password,
			}

			// Make request
			rr := MakeRequest("POST", "/api/auth/register", reqBody, router)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rr.Code)
			}

			// If should succeed, check response structure
			if tc.shouldSucceed {
				var resp models.AuthResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
				}

				// Check response fields
				if resp.Token == "" {
					t.Error("Expected token in response, got empty string")
				}
				if resp.UserID <= 0 {
					t.Errorf("Expected positive user ID, got %d", resp.UserID)
				}
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	// Skip if no test database connection
	if TestDB == nil {
		t.Skip("No test database connection")
	}

	// Setup test router
	router := SetupTestRouter(TestDB)

	// Create a test user first
	username := "loginuser"
	password := "loginpassword"
	
	// Register user
	reqBody := models.AuthRequest{
		Username: username,
		Password: password,
	}
	
	registerRR := MakeRequest("POST", "/api/auth/register", reqBody, router)
	if registerRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create test user: %s", registerRR.Body.String())
	}

	// Test cases
	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		shouldSucceed  bool
	}{
		{"ValidLogin", username, password, http.StatusOK, true},
		{"WrongPassword", username, "wrongpassword", http.StatusUnauthorized, false},
		{"NonexistentUser", "nonexistentuser", "password", http.StatusUnauthorized, false},
		{"EmptyUsername", "", password, http.StatusBadRequest, false},
		{"EmptyPassword", username, "", http.StatusBadRequest, false},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request body
			reqBody := models.AuthRequest{
				Username: tc.username,
				Password: tc.password,
			}

			// Make request
			rr := MakeRequest("POST", "/api/auth/login", reqBody, router)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rr.Code)
			}

			// If should succeed, check response structure
			if tc.shouldSucceed {
				var resp models.AuthResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
				}

				// Check response fields
				if resp.Token == "" {
					t.Error("Expected token in response, got empty string")
				}
				if resp.UserID <= 0 {
					t.Errorf("Expected positive user ID, got %d", resp.UserID)
				}
			}
		})
	}
}