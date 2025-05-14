package handlers

import (
	"net/http"
	"time"

	"software-backend/internal/service"

	"github.com/labstack/echo/v4"
)

// Struct to manage dependencies
type BusinessHoursHandler struct {
	service service.BusinessHoursService
}

// Constructor to pass on dependencies
func NewBusinessHoursHandler(service service.BusinessHoursService) *BusinessHoursHandler {
	return &BusinessHoursHandler{service: service}
}

// Get business hours for a specific date
func (h *BusinessHoursHandler) GetBusinessHours(c echo.Context) error {
	// Get date & perform basic input validation
	dateStr := c.QueryParam("date")
	if dateStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing date parameter"})
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format, expected YYYY-MM-DD"})
	}
	// Get intervals for business hours from Service
	intervals, err := h.service.GetBusinessHoursForDate(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, intervals)
}
