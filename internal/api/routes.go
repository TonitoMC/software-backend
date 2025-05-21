package api

import (
	"net/http"

	"software-backend/internal/api/handlers"

	"github.com/labstack/echo/v4"
)

// Set up handlers & other dependencies
type RouterConfig struct {
	AuthHandler          *handlers.AuthHandler
	UserHandler          *handlers.UserHandler
	AppointmentHandler   *handlers.AppointmentHandler
	PatientHandler       *handlers.PatientHandler
	BusinessHoursHandler *handlers.BusinessHoursHandler
}

// Sets up routes for the application
func SetupRoutes(e *echo.Echo, config *RouterConfig) {
	e.POST("/login", config.AuthHandler.Login)
	e.POST("/register", config.UserHandler.Register)

	// Apointment routes, some overlap but will be fixed for later versions
	e.GET("/appointments", config.AppointmentHandler.GetAppointmentsInDateRange)
	e.GET("/appointments/today", config.AppointmentHandler.GetTodaysAppointments)
	e.GET("/appointments/month", config.AppointmentHandler.GetAppointmentsForMonth)
	e.GET("appointments/day", config.AppointmentHandler.GetAppointmentsForDate)
	e.DELETE("/appointments/:id", config.AppointmentHandler.DeleteAppointment)
	e.PUT("/appointments/:id", config.AppointmentHandler.UpdateAppointment)
	e.POST("/appointments", config.AppointmentHandler.CreateAppointment)

	// Patient routes
	e.GET("/patients/search", config.PatientHandler.SearchPatients)
	e.GET("/patients/:id", config.PatientHandler.GetPatient)

	// Business hours routes
	e.GET("/business-hours", config.BusinessHoursHandler.GetBusinessHours)

	// Route just to verify everything's up
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to the Medical App API!")
	})
}
