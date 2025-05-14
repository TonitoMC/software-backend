package api

import (
	"net/http"

	"software-backend/internal/api/handlers"

	"github.com/labstack/echo/v4"
)

// Set up handlers & other dependencies
type RouterConfig struct {
	AuthHandler        *handlers.AuthHandler
	AppointmentHandler *handlers.AppointmentHandler
}

// Sets up routes for the application
func SetupRoutes(e *echo.Echo, config *RouterConfig) {
	// --- Public Routes (no authentication required) ---
	// Example: Login endpoint
	e.POST("/login", config.AuthHandler.Login)
	e.GET("/appointments", config.AppointmentHandler.GetAppointmentsInDateRange)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to the Medical App API!")
	})
}
