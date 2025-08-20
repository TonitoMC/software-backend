package diagnostic

import (
	"software-backend/internal/models"
	"software-backend/internal/repository/diagnostic"
)

type DiagnosticService interface {
	GetByConsultationIDWithTreatments(consultationID int) ([]models.Diagnostic, error)
	CreateBatch(consultationID int, diagnostics []models.Diagnostic) error
}

type diagnosticService struct {
	repo diagnostic.DiagnosticRepository
}

func NewDiagnosticService(repo diagnostic.DiagnosticRepository) DiagnosticService {
	return &diagnosticService{repo: repo}
}

func (s *diagnosticService) GetByConsultationIDWithTreatments(consultationID int) ([]models.Diagnostic, error) {
	return s.repo.GetByConsultationIDWithTreatments(consultationID)
}

func (s *diagnosticService) CreateBatch(consultationID int, diagnostics []models.Diagnostic) error {
	return s.repo.CreateBatch(consultationID, diagnostics)
}
