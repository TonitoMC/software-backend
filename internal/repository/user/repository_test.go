package user

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"software-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestGetUserByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	userID := 1
	expectedUser := &models.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: "hash",
		Email:        "test@example.com",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, username, password_hash, correo FROM usuarios WHERE id = $1`,
	)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "correo"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.PasswordHash, expectedUser.Email),
		)

	user, err := repo.GetUserByID(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *user != *expectedUser {
		t.Errorf("unexpected user: %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	userID := 999

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, username, password_hash, correo FROM usuarios WHERE id = $1`,
	)).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByID(userID)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got: %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetUserByID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	userID := 1

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, username, password_hash, correo FROM usuarios WHERE id = $1`,
	)).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	_, err = repo.GetUserByID(userID)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetUserByUsername_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	username := "testuser"
	expectedUser := &models.User{
		ID:           2,
		Username:     username,
		PasswordHash: "hash2",
		Email:        "test2@example.com",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, username, password_hash, correo FROM usuarios WHERE username = $1`,
	)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "correo"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.PasswordHash, expectedUser.Email),
		)

	user, err := repo.GetUserByUsername(username)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *user != *expectedUser {
		t.Errorf("unexpected user: %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	username := "nouser"

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, username, password_hash, correo FROM usuarios WHERE username = $1`,
	)).
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByUsername(username)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got: %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetUserByUsername_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	username := "testuser"

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, username, password_hash, correo FROM usuarios WHERE username = $1`,
	)).
		WithArgs(username).
		WillReturnError(sql.ErrConnDone)

	_, err = repo.GetUserByUsername(username)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	user := models.User{
		Username:     "newuser",
		PasswordHash: "hash",
		Email:        "new@example.com",
	}
	expectedID := 10

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO usuarios (username, password_hash, correo)
						VALUES ($1, $2, $3)
						RETURNING id`,
	)).
		WithArgs(user.Username, user.PasswordHash, user.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	created, err := repo.CreateUser(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != expectedID ||
		created.Username != user.Username ||
		created.PasswordHash != user.PasswordHash ||
		created.Email != user.Email {
		t.Errorf("unexpected created user: %+v", created)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	user := models.User{
		Username:     "dupeuser",
		PasswordHash: "hash",
		Email:        "dupe@example.com",
	}

	pqErr := &pq.Error{Code: "23505", Message: "duplicate key value violates unique constraint", Severity: "ERROR"}
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO usuarios (username, password_hash, correo)
						VALUES ($1, $2, $3)
						RETURNING id`,
	)).
		WithArgs(user.Username, user.PasswordHash, user.Email).
		WillReturnError(pqErr)

	_, err = repo.CreateUser(user)
	if !errors.Is(err, ErrDuplicateUsername) {
		t.Fatalf("expected ErrDuplicateUsername, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestCreateUser_OtherError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	user := models.User{
		Username:     "failuser",
		PasswordHash: "hash",
		Email:        "fail@example.com",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO usuarios (username, password_hash, correo)
						VALUES ($1, $2, $3)
						RETURNING id`,
	)).
		WithArgs(user.Username, user.PasswordHash, user.Email).
		WillReturnError(sql.ErrConnDone)

	_, err = repo.CreateUser(user)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
