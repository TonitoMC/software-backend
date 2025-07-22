package user

import (
	"errors"
	"testing"

	"software-backend/internal/mocks"
	"software-backend/internal/models"

	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	username := "testuser"
	email := "test@example.com"
	password := "secret"

	// We don't know the hash in advance, so use gomock.Any()
	mockRepo.EXPECT().
		CreateUser(gomock.AssignableToTypeOf(models.User{})).
		DoAndReturn(func(u models.User) (*models.User, error) {
			// Check that the username and email are passed through
			if u.Username != username || u.Email != email {
				t.Errorf("unexpected user: %+v", u)
			}
			// Check that the password is hashed
			if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
				t.Errorf("password not hashed correctly: %v", err)
			}
			u.ID = 1
			return &u, nil
		})

	user, err := svc.RegisterUser(username, email, password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != username || user.Email != email || user.ID != 1 {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestRegisterUser_InvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	_, err := svc.RegisterUser("", "test@example.com", "password")
	if err == nil || !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}

	_, err = svc.RegisterUser("testuser", "test@example.com", "")
	if err == nil || !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestRegisterUser_HashError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	// Expect CreateUser to NOT be called
	mockRepo.EXPECT().CreateUser(gomock.Any()).Times(0)

	// Patch bcrypt
	orig := bcryptGenerateFromPassword
	bcryptGenerateFromPassword = func([]byte, int) ([]byte, error) {
		return nil, errors.New("hash error")
	}
	defer func() { bcryptGenerateFromPassword = orig }()

	svc := NewUserService(mockRepo)

	_, err := svc.RegisterUser("testuser", "test@example.com", "password")
	if err == nil || !errors.Is(err, ErrPasswordHashingFailed) {
		t.Errorf("expected ErrPasswordHashingFailed, got %v", err)
	}
}

func TestRegisterUser_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	username := "testuser"
	email := "test@example.com"
	password := "secret"

	mockRepo.EXPECT().
		CreateUser(gomock.AssignableToTypeOf(models.User{})).
		Return(nil, errors.New("db error"))

	_, err := svc.RegisterUser(username, email, password)
	if err == nil || err.Error() != "failed to create user account" {
		t.Errorf("expected 'failed to create user account', got %v", err)
	}
}
