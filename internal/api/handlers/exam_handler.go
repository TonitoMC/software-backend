package handlers

import (
	"net/http"
	"strconv"

	"software-backend/internal/service"

	"github.com/labstack/echo/v4"
)

// Struct to manage dependencies
type ExamHandler struct {
	service service.ExamService
}

// Constructor to pass on dependencies
func NewExamHandler(service service.ExamService) *ExamHandler {
	return &ExamHandler{service: service}
}

// Get business hours for a specific date
func (h *ExamHandler) GetByPatientID(c echo.Context) error {
	patientID := c.Param("patient_id")
	id, err := strconv.Atoi(patientID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid patient"})
	}
	// Get date & perform basic input validation
	exams, err := h.service.GetByPatientID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, exams)
}
