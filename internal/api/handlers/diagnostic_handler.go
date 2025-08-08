package handlers

import (
	"net/http"
	"strconv"

	"software-backend/internal/models"
	"software-backend/internal/service/diagnostic"

	"github.com/labstack/echo/v4"
)

type DiagnosticHandler struct {
	service diagnostic.DiagnosticService
}

func NewDiagnosticHandler(s diagnostic.DiagnosticService) *DiagnosticHandler {
	return &DiagnosticHandler{service: s}
}

func (h *DiagnosticHandler) GetByConsultationID(c echo.Context) error {
	consultationID, err := strconv.Atoi(c.Param("consultation_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid consultation id"})
	}
	diagnostics, err := h.service.GetByConsultationIDWithTreatments(consultationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, diagnostics)
}

func (h *DiagnosticHandler) CreateBatch(c echo.Context) error {
	consultationID, err := strconv.Atoi(c.Param("consultation_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid consultation id"})
	}

	var diagnostics []models.Diagnostic
	if err := c.Bind(&diagnostics); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if err := h.service.CreateBatch(consultationID, diagnostics); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]string{"status": "created"})
}
