package handlers

import (
	"net/http"
	"strconv"

	service "software-backend/internal/service/consultation"

	"github.com/labstack/echo/v4"
)

type ConsultationHandler struct {
	service service.ConsultationService
}

func NewConsultationHandler(service service.ConsultationService) *ConsultationHandler {
	return &ConsultationHandler{service: service}
}

func (h *ConsultationHandler) GetByPatientID(c echo.Context) error {
	patientID := c.Param("patient_id")
	id, err := strconv.Atoi(patientID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid patient"})
	}
	consultations, err := h.service.GetByPatientID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, consultations)
}
