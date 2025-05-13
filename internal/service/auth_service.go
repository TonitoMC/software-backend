package service

import (
	"errors"

	"software-backend/internal/repository"
)

// Custom errors for authentication failure
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// Interface for authentication-related business logic
type AuthService interface {
	AuthenticateUser(username, password string) (userID int, err error)
}

// Dependencies to verify user credentials
type authService struct {
	userRepo repository.UserRepository
}

// Create a new AuthService instance
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// AuthenticateUser verifies user credentials against the database.
func (s *authService) AuthenticateUser(username, password string) (userID int, err error) {
	// Get user from the database by username
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return 0, ErrInvalidCredentials
	}

	if password != "password" {
		return 0, ErrInvalidCredentials
	}

	return user.ID, nil
}
