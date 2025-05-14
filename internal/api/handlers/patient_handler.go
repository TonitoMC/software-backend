package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"software-backend/internal/repository"
	"software-backend/internal/service"
)

// Struct to manage dependencies
type PatientHandler struct {
	patientService service.PatientService
}

// Constructor for passing on dependencies
func NewPatientHandler(svc service.PatientService) *PatientHandler {
	return &PatientHandler{
		patientService: svc,
	}
}

// Get patient by ID
func (h *PatientHandler) GetPatient(c echo.Context) error {
	// Get id from request
	idStr := c.Param("id")

	// Perform basic input validation / convert to int
	patientID, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid patient ID format")
	}
	if patientID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Patient ID must be positive")
	}

	// Get data from Service
	patient, err := h.patientService.GetPatientByID(patientID)
	if err != nil {
		// Specific error handling
		if errors.Is(err, repository.ErrPatientNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Patient not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve patient") // Generic 500 error
	}

	return c.JSON(http.StatusOK, patient)
}

// Search patients
func (h *PatientHandler) SearchPatients(c echo.Context) error {
	// Get name to match against from param & perform basic validation
	q := c.QueryParam("q")
	if len(q) < 2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "query too short"})
	}
	// Get list of patients from Service
	patients, err := h.patientService.SearchPatients(q, 10)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, patients)
}
