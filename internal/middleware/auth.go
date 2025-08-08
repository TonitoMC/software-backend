package middleware

import (
	"net/http"
	"strings"
	"time"

	"software-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte("your-secret-key")

func GenerateToken(user *models.User, roles []string) (string, error) {
	claims := &models.JWTClaims{
		UserID:   uint(user.ID),
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}

			tokenString := strings.Replace(auth, "Bearer ", "", 1)

			token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{},
				func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims := token.Claims.(*models.JWTClaims)
			c.Set("user_id", int(claims.UserID))
			c.Set("username", claims.Username)
			c.Set("roles", claims.Roles)

			return next(c)
		}
	}
}

func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoles, ok := c.Get("roles").([]string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "no roles found")
			}

			// Check if user has any of the allowed roles
			for _, userRole := range userRoles {
				for _, allowedRole := range allowedRoles {
					if userRole == allowedRole {
						return next(c)
					}
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

// Helper function to check if user has specific role
func HasRole(c echo.Context, role string) bool {
	userRoles, ok := c.Get("roles").([]string)
	if !ok {
		return false
	}

	for _, userRole := range userRoles {
		if userRole == role {
			return true
		}
	}
	return false
}
