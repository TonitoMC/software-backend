package handlers

import (
	"net/http"

	"software-backend/internal/middleware"
	service "software-backend/internal/service/auth"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService service.AuthService
	jwtSecret   string
}

func NewAuthHandler(svc service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: svc,
		jwtSecret:   jwtSecret,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type UserInfo struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid login request body")
	}

	// Use email if provided, otherwise use username
	identifier := req.Email
	if identifier == "" {
		identifier = req.Username
	}

	// Auth service verifies credentials
	user, err := h.authService.AuthenticateUser(identifier, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// Get user roles (you'll need to implement this in your auth service)
	roles, err := h.authService.GetUserRoles(user.ID)
	if err != nil {
		// Default to basic user role if no roles found
		roles = []string{"user"}
	}

	// Create JWT with roles
	token, err := middleware.GenerateToken(user, roles)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    roles,
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  userInfo,
	})
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	_, ok = c.Get("username").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username")
	}

	roles, ok := c.Get("roles").([]string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid roles")
	}

	// Optionally fetch fresh user data
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    roles,
	}

	return c.JSON(http.StatusOK, userInfo)
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	// Get fresh user data and roles
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	roles, err := h.authService.GetUserRoles(userID)
	if err != nil {
		roles = []string{"user"}
	}

	token, err := middleware.GenerateToken(user, roles)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}

	return c.JSON(http.StatusOK, LoginResponse{Token: token})
}
