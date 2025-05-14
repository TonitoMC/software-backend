package service

import (
	"errors"

	"software-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// Custom errors for authentication failure, probably gonna be moved
// to a separate file in the future
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// Interface holds the expected methods from the service
type AuthService interface {
	AuthenticateUser(username, password string) (userID int, err error)
}

// Struct to manage dependencies
type authService struct {
	userRepo repository.UserRepository
}

// Constructor to pass on dependencies
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// Verify user credentials against database
func (s *authService) AuthenticateUser(username, password string) (userID int, err error) {
	// Get user by username via repository
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return 0, ErrInvalidCredentials
	}

	// Compare password using BCrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, errors.New("internal authentication error")
	}

	return user.ID, nil
}
