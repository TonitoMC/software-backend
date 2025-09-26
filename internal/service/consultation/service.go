package consultation

import (
	"errors"
	"fmt"
	"time"

	"software-backend/internal/models"
	"software-backend/internal/repository/consultation"
	"software-backend/internal/repository/diagnostic"
	questionnaire "software-backend/internal/service/questionnaire"
)

type ConsultationService interface {
	// Existing operation
	GetByPatientID(patientID int) ([]models.Consultation, error)

	// New operations
	Create(req CreateConsultationRequest) (*models.Consultation, error)
	GetByID(id int) (*models.Consultation, error)
	GetWithDetails(id int) (*models.CompleteConsultation, error)
	Update(id int, req UpdateConsultationRequest) (*models.Consultation, error)
	Delete(id int) error
}

type consultationService struct {
	repo                 consultation.ConsultationRepository
	diagnosticRepo       diagnostic.DiagnosticRepository
	questionnaireService questionnaire.QuestionnaireService
}

func NewConsultationService(
	repo consultation.ConsultationRepository,
	diagnosticRepo diagnostic.DiagnosticRepository,
	questionnaireService questionnaire.QuestionnaireService,
) ConsultationService {
	return &consultationService{
		repo:                 repo,
		diagnosticRepo:       diagnosticRepo,
		questionnaireService: questionnaireService,
	}
}

// Existing method - unchanged
func (s *consultationService) GetByPatientID(patientID int) ([]models.Consultation, error) {
	return s.repo.GetByPatientID(patientID)
}

// New methods
func (s *consultationService) Create(req CreateConsultationRequest) (*models.Consultation, error) {
	// Validate questionnaire exists if provided
	if req.QuestionnaireID != nil {
		if err := s.questionnaireService.ValidateQuestionnaireExists(*req.QuestionnaireID); err != nil {
			return nil, errors.New("questionnaire not found")
		}
	}

	// Validate required fields
	if req.PatientID <= 0 {
		return nil, errors.New("patient ID is required")
	}
	if req.Reason == "" {
		return nil, errors.New("reason is required")
	}

	consultation := models.Consultation{
		PatientID:       req.PatientID,
		QuestionnaireID: req.QuestionnaireID,
		Reason:          req.Reason,
		Date:            time.Now(),
	}

	if !req.Date.IsZero() {
		consultation.Date = req.Date
	}

	id, err := s.repo.Create(consultation)
	if err != nil {
		return nil, err
	}

	consultation.ID = id
	return &consultation, nil
}

func (s *consultationService) GetByID(id int) (*models.Consultation, error) {
	if id <= 0 {
		return nil, errors.New("invalid consultation ID")
	}
	return s.repo.GetByID(id)
}

func (s *consultationService) Update(id int, req UpdateConsultationRequest) (*models.Consultation, error) {
	// Validate consultation exists
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Validate questionnaire if provided
	if req.QuestionnaireID != nil {
		if err := s.questionnaireService.ValidateQuestionnaireExists(*req.QuestionnaireID); err != nil {
			return nil, errors.New("questionnaire not found")
		}
	}

	// Update fields if provided
	if req.Reason != "" {
		existing.Reason = req.Reason
	}
	if !req.Date.IsZero() {
		existing.Date = req.Date
	}
	if req.QuestionnaireID != nil {
		existing.QuestionnaireID = req.QuestionnaireID
	}

	err = s.repo.Update(id, *existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *consultationService) Delete(id int) error {
	// Check if consultation exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}

// Request/Response types
type CreateConsultationRequest struct {
	PatientID       int       `json:"patient_id" validate:"required"`
	QuestionnaireID *int      `json:"questionnaire_id,omitempty"`
	Reason          string    `json:"reason" validate:"required"`
	Date            time.Time `json:"date,omitempty"`
}

type UpdateConsultationRequest struct {
	Reason          string    `json:"reason,omitempty"`
	QuestionnaireID *int      `json:"questionnaire_id,omitempty"`
	Date            time.Time `json:"date,omitempty"`
}

func (s *consultationService) GetWithDetails(id int) (*models.CompleteConsultation, error) {
	fmt.Printf("DEBUG: Service GetWithDetails called with ID: %d\n", id)

	// Get consultation with questions
	complete, err := s.repo.GetComplete(id)
	if err != nil {
		fmt.Printf("DEBUG: Error from repo.GetComplete: %v\n", err)
		return nil, err
	}

	fmt.Printf("DEBUG: Got complete consultation, getting diagnostics...\n")

	// Get diagnostics using your existing diagnostic repo
	diagnostics, err := s.diagnosticRepo.GetByConsultationIDWithTreatments(id)
	if err != nil {
		fmt.Printf("DEBUG: Error from diagnosticRepo: %v\n", err)
		return nil, err
	}

	complete.Diagnoses = diagnostics
	return complete, nil
}
