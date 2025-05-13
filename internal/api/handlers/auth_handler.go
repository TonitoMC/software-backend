package handlers

import (
	"net/http"
	"time"

	"software-backend/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Auth Handler dependencies
type AuthHandler struct {
	authService service.AuthService
	jwtSecret   string
}

// Creates a new AuthHandler instance
func NewAuthHandler(svc service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: svc,
		jwtSecret:   jwtSecret,
	}
}

// Body of a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Body of a login response
type LoginResponse struct {
	Token string `json:"token"`
}

// Login handles POST requests to the login endpoint
func (h *AuthHandler) Login(c echo.Context) error {
	// Parse the body of the request
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
