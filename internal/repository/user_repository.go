package repository

import (
	"errors"
	"sync"

	"software-backend/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

// Interface for user data access operations
type UserRepository interface {
	GetUserByID(id int) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
}

type MockUserRepository struct {
	users      map[int]*models.User
	mu         sync.Mutex
	nextUserID int
}

func NewMockUserRepository(initialUsers map[int]*models.User) UserRepository {
	if initialUsers == nil {
		initialUsers = make(map[int]*models.User)
	}
	nextID := 1
	for id := range initialUsers {
		if id >= nextID {
			nextID = id + 1
		}
	}

	return &MockUserRepository{
		users:      initialUsers,
		nextUserID: nextID,
	}
}

func (r *MockUserRepository) GetUserByID(id int) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	copiedUser := *user
	return &copiedUser, nil
}

func (r *MockUserRepository) GetUserByUsername(username string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, user := range r.users {
		if user.Username == username {
			copiedUser := *user
			return &copiedUser, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *MockUserRepository) CreateUser(user models.User) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user.ID = r.nextUserID
	r.nextUserID++
	r.users[user.ID] = &user
	createdUser := user
	return &createdUser, nil
}
