package auth

import (
	"errors"
	"testing"

	"software-backend/internal/mocks"
	"software-backend/internal/models"

	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthenticateUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	password := "secret"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: string(hash),
	}

	mockRepo.EXPECT().
		GetUserByUsername("testuser").
		Return(user, nil)

	svc := NewAuthService(mockRepo)
	userID, err := svc.AuthenticateUser("testuser", password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != user.ID {
		t.Errorf("expected userID %d, got %d", user.ID, userID)
	}
}

func TestAuthenticateUser_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRepo.EXPECT().
		GetUserByUsername("nouser").
		Return(nil, errors.New("not found"))

	svc := NewAuthService(mockRepo)
	_, err := svc.AuthenticateUser("nouser", "irrelevant")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticateUser_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &models.User{
		ID:           2,
		Username:     "testuser",
		PasswordHash: string(hash),
	}

	mockRepo.EXPECT().
		GetUserByUsername("testuser").
		Return(user, nil)

	svc := NewAuthService(mockRepo)
	_, err := svc.AuthenticateUser("testuser", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
