package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"software-backend/internal/models"
	service "software-backend/internal/service/consultation"

	"github.com/labstack/echo/v4"
)

type ConsultationHandler struct {
	service service.ConsultationService
}

func NewConsultationHandler(service service.ConsultationService) *ConsultationHandler {
	return &ConsultationHandler{service: service}
}

// ----------------------------------------------------
// Existing endpoints
// ----------------------------------------------------

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

func (h *ConsultationHandler) Create(c echo.Context) error {
	var req service.CreateConsultationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	consultation, err := h.service.Create(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, consultation)
}

func (h *ConsultationHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid consultation ID"})
	}

	consultation, err := h.service.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "consultation not found"})
	}

	return c.JSON(http.StatusOK, consultation)
}

func (h *ConsultationHandler) GetWithDetails(c echo.Context) error {
	fmt.Println("DEBUG: GetWithDetails handler called")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid consultation ID"})
	}

	consultation, err := h.service.GetWithDetails(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "consultation not found"})
	}

	return c.JSON(http.StatusOK, consultation)
}

func (h *ConsultationHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid consultation ID"})
	}

	var req service.UpdateConsultationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	consultation, err := h.service.Update(id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, consultation)
}

func (h *ConsultationHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid consultation ID"})
	}

	err = h.service.Delete(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusNoContent, nil)
}

// ----------------------------------------------------
// New endpoint: CreateFromQuestionnaire
// ----------------------------------------------------
//
// Receives a JSON body like:
//
//	{
//	  "patientId": 1,
//	  "questionnaireId": 2,
//	  "date": "2025-10-23T20:30:58.482Z",
//	  "answers": { "1": { "od": "...", "oi": "...", "comment": "..." } }
//	}
func (h *ConsultationHandler) CreateFromQuestionnaire(c echo.Context) error {
	var req models.ConsultationFromQuestionnaireRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	consultation, err := h.service.CreateFromQuestionnaire(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, consultation)
}
