package services

import (
	"errors"

	"officestonks/internal/auth"
	"officestonks/internal/models"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo models.UserRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo models.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// Register creates a new user account
func (s *AuthService) Register(username, password string) (*models.AuthResponse, error) {
	// Check if username already exists
	_, err := s.userRepo.GetUserByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}
	
	// Hash the password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	
	// Create the user
	user, err := s.userRepo.CreateUser(username, hashedPassword)
	if err != nil {
		return nil, err
	}
	
	// Generate a JWT token
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}
	
	// Return the auth response
	return &models.AuthResponse{
		Token:   token,
		UserID:  user.ID,
		IsAdmin: user.IsAdmin,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(username, password string) (*models.AuthResponse, error) {
	// Get the user by username
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}
	
	// Verify the password
	valid, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return nil, errors.New("invalid username or password")
	}
	
	// Generate a JWT token
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}
	
	// Return the auth response
	return &models.AuthResponse{
		Token:   token,
		UserID:  user.ID,
		IsAdmin: user.IsAdmin,
	}, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *AuthService) ValidateToken(tokenString string) (int, error) {
	// Validate the token
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	
	// Check if the user exists
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return 0, errors.New("invalid token: user not found")
	}
	
	return user.ID, nil
}