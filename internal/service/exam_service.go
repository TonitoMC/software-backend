// service/business_hours_service.go
package service

import (
	"software-backend/internal/models"
	"software-backend/internal/repository"
)

// BussinessHoursService interface defines the methods expected from the service
type ExamService interface {
	GetByPatientID(patientID int) ([]models.Exam, error)
}

// Struct to manage dependencies
type examService struct {
	repo repository.ExamRepository
}

// Constructor to pass on dependencies
func NewExamService(repo repository.ExamRepository) ExamService {
	return &examService{repo: repo}
}

func (s *examService) GetByPatientID(patientID int) ([]models.Exam, error) {
	return s.repo.GetByPatientID(patientID)
}
