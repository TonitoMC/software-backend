package auth

import (
	"errors"

	"software-backend/internal/models"
	"software-backend/internal/repository/user"

	"golang.org/x/crypto/bcrypt"
)

// Custom errors for authentication failure
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// Interface holds the expected methods from the service
type AuthService interface {
	AuthenticateUser(username, password string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	GetUserRoles(userID int) ([]string, error)
}

// Struct to manage dependencies
type authService struct {
	userRepo user.UserRepository
}

// Constructor to pass on dependencies
func NewAuthService(userRepo user.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// Verify user credentials against database and return full user
func (s *authService) AuthenticateUser(usernameOrEmail, password string) (*models.User, error) {
	// Try to get user by username first
	user, err := s.userRepo.GetUserByUsername(usernameOrEmail)
	if err != nil {
		// If not found by username, try by email
		user, err = s.userRepo.GetUserByEmail(usernameOrEmail)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	// Compare password using BCrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, errors.New("internal authentication error")
	}

	return user, nil
}

// Get user by ID
func (s *authService) GetUserByID(id int) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Get user roles - implement based on your role system
func (s *authService) GetUserRoles(userID int) ([]string, error) {
	// Option 1: Static role assignment (simple approach)
	// You can make this dynamic by adding a roles table later

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// For now, assign roles based on some logic
	// You can make this more sophisticated later
	roles := []string{"user"} // Default role

	// Example: hardcoded admin users (replace with your logic)
	adminUsernames := []string{"admin", "administrator"} // or check by ID, email, etc.
	for _, adminUsername := range adminUsernames {
		if user.Username == adminUsername {
			roles = append(roles, "admin")
			break
		}
	}

	return roles, nil
}
