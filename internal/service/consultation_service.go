// service/business_hours_service.go
package service

import (
	"software-backend/internal/models"
	"software-backend/internal/repository"
)

type ConsultationService interface {
	GetByPatientID(patientID int) ([]models.Consultation, error)
}

type consultationService struct {
	repo repository.ConsultationRepository
}

func NewConsultationService(repo repository.ConsultationRepository) ConsultationService {
	return &consultationService{repo: repo}
}

func (s *consultationService) GetByPatientID(patientID int) ([]models.Consultation, error) {
	return s.repo.GetByPatientID(patientID)
}
