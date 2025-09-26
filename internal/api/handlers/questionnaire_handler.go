package handlers

import (
	"net/http"
	"strconv"

	"software-backend/internal/models"
	service "software-backend/internal/service/questionnaire"

	"github.com/labstack/echo/v4"
)

type QuestionnaireHandler struct {
	service service.QuestionnaireService
}

func NewQuestionnaireHandler(service service.QuestionnaireService) *QuestionnaireHandler {
	return &QuestionnaireHandler{service: service}
}

func (h *QuestionnaireHandler) GetActive(c echo.Context) error {
	questionnaires, err := h.service.GetActiveQuestionnaires()
	if err != nil {
		c.Logger().Error("Error getting active questionnaires: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

	}
	return c.JSON(http.StatusOK, questionnaires)
}

func (h *QuestionnaireHandler) GetWithQuestions(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid questionnaire ID"})
	}

	questionnaire, err := h.service.GetQuestionnaireWithQuestions(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "questionnaire not found"})
	}

	return c.JSON(http.StatusOK, questionnaire)
}

// Add these methods to the existing QuestionnaireHandler struct

func (h *QuestionnaireHandler) GetAll(c echo.Context) error {
	questionnaires, err := h.service.GetAllQuestionnaires()
	if err != nil {
		c.Logger().Error("Error getting all questionnaires: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, questionnaires)
}

func (h *QuestionnaireHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid questionnaire ID"})
	}

	var questionnaire models.QuestionnaireUpdate
	if err := c.Bind(&questionnaire); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := h.service.UpdateQuestionnaire(id, &questionnaire); err != nil {
		c.Logger().Error("Error updating questionnaire: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "questionnaire updated successfully"})
}

func (h *QuestionnaireHandler) SetActive(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid questionnaire ID"})
	}

	var request struct {
		Active bool `json:"active"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := h.service.SetQuestionnaireActive(id, request.Active); err != nil {
		c.Logger().Error("Error setting questionnaire active status: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "questionnaire status updated successfully"})
}
