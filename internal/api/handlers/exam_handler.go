package handlers

import (
	"net/http"
	"strconv"

	examservice "software-backend/internal/service/exam"

	"github.com/labstack/echo/v4"
)

type ExamHandler struct {
	service examservice.ExamService
}

func NewExamHandler(service examservice.ExamService) *ExamHandler {
	return &ExamHandler{service: service}
}

// GET /api/patients/:patientId/exams
func (h *ExamHandler) GetExamsByPatient(c echo.Context) error {
	patientID, err := strconv.Atoi(c.Param("patientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid patient ID",
		})
	}

	exams, err := h.service.GetByPatientID(patientID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch exams",
		})
	}

	return c.JSON(http.StatusOK, exams)
}

// POST /api/exams/:examId/upload
func (h *ExamHandler) UploadPDF(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid exam ID",
		})
	}

	file, err := c.FormFile("pdf")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No PDF file provided",
		})
	}

	err = h.service.UploadPDF(examID, file)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "File uploaded successfully",
	})
}

// GET /api/exams/:examId/download
func (h *ExamHandler) GetDownloadURL(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid exam ID",
		})
	}

	url, err := h.service.GetDownloadURL(examID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"download_url": url,
	})
}
