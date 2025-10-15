package user

import (
	"database/sql"
	"errors"
	"fmt"

	"software-backend/internal/models"

	"github.com/lib/pq"
)

// Custom errors, probably moved to another file later on
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicateUsername = errors.New("duplicate username")
)

// Interface defines the methods for interaction with the Repository
type UserRepository interface {
	GetUserByID(id int) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user models.User) (*models.User, error)
}

// Struct to manage dependencies
type userRepository struct {
	db *sql.DB
}

// Constructor to pass on dependencies
func NewUserRepository(dbConn *sql.DB) UserRepository {
	return &userRepository{
		db: dbConn,
	}
}

// Get a User via their ID
func (r *userRepository) GetUserByID(id int) (*models.User, error) {
	// Build query
	query := `SELECT id, username, password_hash, correo FROM usuarios WHERE id = $1`

	// Create model
	user := &models.User{}

	// Extract query into user, if unable to return error
	if err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to get user by ID %d: %w", id, err)
	}
	return user, nil
}

// Get a user by name
func (r *userRepository) GetUserByUsername(username string) (*models.User, error) {
	// Build query
	query := `SELECT id, username, password_hash, correo FROM usuarios WHERE username = $1`

	// Create model
	user := &models.User{}

	// Extract query into user, if unable return error
	if err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to get user by username %s: %w", username, err)
	}

	return user, nil
}

// Get a user by email
func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	// Build query
	query := `SELECT id, username, password_hash, correo FROM usuarios WHERE correo = $1`

	// Create model
	user := &models.User{}

	// Extract query into user, if unable return error
	if err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to get user by email %s: %w", email, err)
	}

	return user, nil
}

// Create a user from a User model
func (r *userRepository) CreateUser(user models.User) (*models.User, error) {
	// Build query
	query := `INSERT INTO usuarios (username, password_hash, correo)
						VALUES ($1, $2, $3)
						RETURNING id`

	var userID int

	// Exec query, return user ID on success or error on failure
	err := r.db.QueryRow(query, user.Username, user.PasswordHash, user.Email).Scan(&userID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code.Name() == "unique_violation" {
				return nil, ErrDuplicateUsername
			}
		}
		return nil, fmt.Errorf("repository: failed to create user: %w", err)
	}

	user.ID = userID

	return &user, nil
}
