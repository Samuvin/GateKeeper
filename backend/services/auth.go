package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"GateKeeper/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication-related business logic
type AuthService struct {
	// In a real microservices architecture, this would be a database client
	// For demonstration, we'll use an in-memory store
	users map[string]*models.User
}

// NewAuthService creates a new authentication service
func NewAuthService() *AuthService {
	return &AuthService{
		users: make(map[string]*models.User),
	}
}

// CreateUser creates a new user with the provided details
func (s *AuthService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error) {
	// Check if user already exists
	if _, exists := s.users[req.Email]; exists {
		return nil, errors.New("user with this email already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &models.User{
		ID:        len(s.users) + 1, // Simple ID generation for demo
		Email:     req.Email,
		Username:  req.Username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// Store user (in real app, this would be saved to database)
	s.users[req.Email] = user

	// Return user response (without password)
	response := user.ToResponse()
	return &response, nil
}

// LoginUser authenticates a user with email and password
func (s *AuthService) LoginUser(ctx context.Context, req models.LoginRequest) (*models.UserResponse, error) {
	// Find user by email
	user, exists := s.users[req.Email]
	if !exists {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Return user response
	response := user.ToResponse()
	return &response, nil
}

// GetUserByEmail retrieves a user by their email address
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*models.UserResponse, error) {
	user, exists := s.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}

	response := user.ToResponse()
	return &response, nil
}

// GetAllUsers returns all users (for demo purposes)
func (s *AuthService) GetAllUsers(ctx context.Context) ([]models.UserResponse, error) {
	var users []models.UserResponse
	for _, user := range s.users {
		users = append(users, user.ToResponse())
	}
	return users, nil
}
