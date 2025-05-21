package handlers

import (
	"net/http"
	"time"

	"software-backend/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Struct to manage dependencies
type AuthHandler struct {
	authService service.AuthService
	jwtSecret   string
}

// Constructor to pass on dependencies
func NewAuthHandler(svc service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: svc,
		jwtSecret:   jwtSecret,
	}
}

// Req for login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Response for login
type LoginResponse struct {
	Token string `json:"token"`
}

// Handle login requests
func (h *AuthHandler) Login(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid login request body")
	}

	// Auth service verifies credentials
	userID, err := h.authService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// If auth is successful, create a JWT
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}

	return c.JSON(http.StatusOK, LoginResponse{Token: signedToken})
}
