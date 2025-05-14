package handlers

import (
	"errors"
	"net/http"

	"software-backend/internal/repository"
	"software-backend/internal/service"

	"github.com/labstack/echo/v4"
)

// Struct to manage dependencies
type UserHandler struct {
	userService service.UserService
}

// Constructor to pass on dependencies
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		userService: svc,
	}
}

// Register request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register a new user
func (h *UserHandler) Register(c echo.Context) error {
	// Bind payload to request
	req := new(RegisterRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid registration request body")
	}

	// Register the user via Service
	createdUser, err := h.userService.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		if errors.Is(err, repository.ErrDuplicateUsername) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		if errors.Is(err, service.ErrPasswordHashingFailed) {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process password")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	return c.JSON(http.StatusCreated, createdUser)
}
