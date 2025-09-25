package handlers

import (
	"errors"
	"net/http"

	repository "software-backend/internal/repository/user"
	service "software-backend/internal/service/user"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		userService: svc,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	Message  string   `json:"message"`
}

func (h *UserHandler) Register(c echo.Context) error {
	req := new(RegisterRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid registration request body")
	}

	// Normalize will happen in service, but light trim here is okay too
	createdUser, err := h.userService.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, repository.ErrDuplicateUsername) {
			return echo.NewHTTPError(http.StatusConflict, "email already exists")
		}
		if errors.Is(err, service.ErrPasswordHashingFailed) {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	// Assign default role to new user
	defaultRoles := []string{"user"}

	response := RegisterResponse{
		ID:       createdUser.ID,
		Username: createdUser.Username,
		Email:    createdUser.Email,
		Roles:    defaultRoles,
		Message:  "User registered successfully",
	}

	return c.JSON(http.StatusCreated, response)
}
