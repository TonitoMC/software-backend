package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"software-backend/internal/models"
	"software-backend/internal/service"

	"github.com/labstack/echo/v4"
)

// Struct to manage dependencies
type AppointmentHandler struct {
	appointmentService service.AppointmentService
}

// Constructor to pass on dependencies
func NewAppointmentHandler(svc service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: svc,
	}
}

// Create an appointment
func (h *AppointmentHandler) CreateAppointment(c echo.Context) error {
	// Bind payload to appointment
	var appt models.Appointment
	if err := c.Bind(&appt); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	// Create appointment via Service
	created, err := h.appointmentService.CreateAppointment(appt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, created)
}

// Update an appointment
func (h *AppointmentHandler) UpdateAppointment(c echo.Context) error {
	// Get ID
	idStr := c.Param("id")

	// Basic input validation
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid appointment ID"})
	}

	// Bind payload to appointment
	var appt models.Appointment
	if err := c.Bind(&appt); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	appt.ID = id

	// Update appointment via Service
	if err := h.appointmentService.UpdateAppointment(appt); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// Get appointments within a time range
func (h *AppointmentHandler) GetAppointmentsInDateRange(c echo.Context) error {
	// Parse request params
	startTimeStr := c.QueryParam("start_time")
	endTimeStr := c.QueryParam("end_time")

	// Basic input validation
	if startTimeStr == "" || endTimeStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required 'start_time' or 'end_time' query parameter")
	}

	// Parse the time strings.
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid 'start_time' format: %v", err))
	}
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid 'end_time' format: %v", err))
	}

	// Basic validation
	if !startTime.Before(endTime) {
		return echo.NewHTTPError(http.StatusBadRequest, "'start_time' must be before 'end_time'")
	}

	// Get grouped appointments via Service
	appointmentsGrouped, err := h.appointmentService.GetAppointmentsInDateRangeAndGroupedByDay(startTime, endTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve appointments") // Generic error for API
	}

	return c.JSON(http.StatusOK, appointmentsGrouped)
}

// Get appointments for current day, probably going to be replaced
// TODO check above ^
func (h *AppointmentHandler) GetTodaysAppointments(c echo.Context) error {
	// Call the service method for today's appointments
	appointmentsGrouped, err := h.appointmentService.GetTodaysAppointments()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve today's appointments")
	}

	// Return the grouped results
	return c.JSON(http.StatusOK, appointmentsGrouped)
}

// Get appointments for a specific month via year & month params
func (h *AppointmentHandler) GetAppointmentsForMonth(c echo.Context) error {
	// Get params
	yearStr := c.QueryParam("year")
	monthStr := c.QueryParam("month")

	// Basic input validation
	if yearStr == "" || monthStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required 'year' or 'month' query parameter")
	}

	// Input validation again
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid year format")
	}
	monthInt, err := strconv.Atoi(monthStr)
	if err != nil || monthInt < 1 || monthInt > 12 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid month format or value (1-12)")
	}
	month := time.Month(monthInt)

	// Get appointments via Service
	appointmentsGrouped, err := h.appointmentService.GetAppointmentsForMonth(year, month)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve appointments for month")
	}

	// Return the grouped results
	return c.JSON(http.StatusOK, appointmentsGrouped)
}

// Get appointments for a specific date
func (h *AppointmentHandler) GetAppointmentsForDate(c echo.Context) error {
	// Get date & perform basic input validation
	dateStr := c.QueryParam("date")
	if dateStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing date parameter"})
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format, expected YYYY-MM-DD"})
	}
	// Get appointments via Service
	appts, err := h.appointmentService.GetAppointmentsForDate(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, appts)
}

// Delete an appointment via ID
func (h *AppointmentHandler) DeleteAppointment(c echo.Context) error {
	// Get ID & perform basic input validation
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid appointment ID"})
	}
	// Delete appointment via Service
	err = h.appointmentService.DeleteAppointment(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
