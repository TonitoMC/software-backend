package user

import (
	"errors"
	"strings"

	"software-backend/internal/models"
	repository "software-backend/internal/repository/user"

	"golang.org/x/crypto/bcrypt"
)

var bcryptGenerateFromPassword = bcrypt.GenerateFromPassword

// Custom errors, will probably be moved to another file later on
var (
	ErrPasswordHashingFailed = errors.New("failed to hash password")
	ErrInvalidInput          = errors.New("invalid input data")
)

// Interface UserService defines methods expected from the service
type UserService interface {
	RegisterUser(username, email, password string) (*models.User, error)
}

// Struct to manage dependencies
type userService struct {
	userRepo repository.UserRepository
}

// Constructor to pass on dependencies
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// Method to register a new user
func (s *userService) RegisterUser(username, email, password string) (*models.User, error) {
	// Normalize inputs
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)

	// Basic input validation
	if username == "" || password == "" || email == "" {
		return nil, ErrInvalidInput
	}

	// Check duplicate by email
	if existing, err := s.userRepo.GetUserByEmail(email); err == nil && existing != nil && existing.ID > 0 {
		return nil, repository.ErrDuplicateUsername // reuse duplicate error; consider new ErrDuplicateEmail
	}

	// Hash password for storage, Bcrypt adds salt
	hashedPassword, err := bcryptGenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrPasswordHashingFailed
	}

	// Create User model to hold the information
	newUser := models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
	}

	// Create User in DB via repository
	createdUser, err := s.userRepo.CreateUser(newUser)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateUsername) {
			return nil, repository.ErrDuplicateUsername
		}
		return nil, errors.New("failed to create user account")
	}

	return createdUser, nil
}
