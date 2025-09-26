package handlers

import (
	"net/http"
	"strconv"

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
