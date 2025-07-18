// service/business_hours_service.go
package consultation

import (
	"software-backend/internal/models"
	"software-backend/internal/repository/consultation"
)

type ConsultationService interface {
	GetByPatientID(patientID int) ([]models.Consultation, error)
}

type consultationService struct {
	repo consultation.ConsultationRepository
}

func NewConsultationService(repo consultation.ConsultationRepository) ConsultationService {
	return &consultationService{repo: repo}
}

func (s *consultationService) GetByPatientID(patientID int) ([]models.Consultation, error) {
	return s.repo.GetByPatientID(patientID)
}
