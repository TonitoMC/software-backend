package handlers

import (
	"fmt"
	"net/http"
	"time" // Needed for parsing time from query params

	"github.com/labstack/echo/v4"

	"software-backend/internal/service" // Dependency on AppointmentService
	// You might need to import a package for custom API response types if you define them
	// "your_module_name/internal/api"
)

// AppointmentHandler holds dependencies for appointment-related API handlers.
type AppointmentHandler struct {
	appointmentService service.AppointmentService // Dependency: the appointment service interface
	// You might also need a logger or validation dependency here
	// logger *log.Logger
	// validator validation.ValidationService
}

// NewAppointmentHandler creates a new AppointmentHandler instance.
// It injects the AppointmentService dependency.
func NewAppointmentHandler(svc service.AppointmentService /*, other dependencies */) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: svc,
	}
}

// GetAppointmentsInDateRange handles requests to get appointments within a date range.
// It expects 'start_time' and 'end_time' query parameters in RFC3339 format.
// e.g., GET /appointments?start_time=2023-10-26T00:00:00Z&end_time=2023-10-27T00:00:00Z
func (h *AppointmentHandler) GetAppointmentsInDateRange(c echo.Context) error {
	// 1. Parse Request Parameters (get start_time and end_time from query string)
	startTimeStr := c.QueryParam("start_time")
	endTimeStr := c.QueryParam("end_time")

	// Validate presence of required parameters
	if startTimeStr == "" || endTimeStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required 'start_time' or 'end_time' query parameter")
	}

	// Parse the time strings. Use a standard format like RFC3339 for robustness.
	// Example format string "2006-01-02T15:04:05Z07:00" or time.RFC3339
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid 'start_time' format: %v", err))
	}
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid 'end_time' format: %v", err))
	}

	// Optional: Basic validation that start is before end
	if !startTime.Before(endTime) {
		return echo.NewHTTPError(http.StatusBadRequest, "'start_time' must be before 'end_time'")
	}

	// 2. Call the Service to get grouped appointments
	// Assuming your service method handles potential issues and returns a map or error
	appointmentsGrouped, err := h.appointmentService.GetAppointmentsInDateRangeAndGroupedByDay(startTime, endTime)
	if err != nil {
		// Handle service-level errors
		// Log the error internally: log.Printf("Service error getting appointments: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve appointments") // Generic error for API
	}

	// 3. Format and Send Response
	// The service already returned the data in the desired grouped map format.
	// Return it directly as JSON.
	return c.JSON(http.StatusOK, appointmentsGrouped)

	// If you had custom API response types in internal/api/, you would map
	// the service result (map[string][]models.Appointment) to your
	// internal/api/AppointmentsGroupedByDateResponse type here before c.JSON().
}

// --- Add other appointment-related handlers here ---
// func (h *AppointmentHandler) CreateAppointment(c echo.Context) error { ... }
// func (h *AppointmentHandler) GetAppointment(c echo.Context) error { ... }
// ... and so on for all relevant endpoints.
