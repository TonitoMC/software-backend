package consultation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"software-backend/internal/models"
	"software-backend/internal/repository/consultation"
	"software-backend/internal/repository/diagnostic"
	questionnaireService "software-backend/internal/service/questionnaire"
)

type ConsultationService interface {
	GetByPatientID(patientID int) ([]models.Consultation, error)
	Create(req CreateConsultationRequest) (*models.Consultation, error)
	GetByID(id int) (*models.Consultation, error)
	GetWithDetails(id int) (*models.CompleteConsultation, error)
	Update(id int, req UpdateConsultationRequest) (*models.Consultation, error)
	Delete(id int) error

	// New ERD-based operation
	CreateFromQuestionnaire(req models.ConsultationFromQuestionnaireRequest) (*models.Consultation, error)
}

type consultationService struct {
	repo                 consultation.ConsultationRepository
	diagnosticRepo       diagnostic.DiagnosticRepository
	questionnaireService questionnaireService.QuestionnaireService
}

func NewConsultationService(
	repo consultation.ConsultationRepository,
	diagnosticRepo diagnostic.DiagnosticRepository,
	questionnaireService questionnaireService.QuestionnaireService,
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

// Create basic consultation
func (s *consultationService) Create(req CreateConsultationRequest) (*models.Consultation, error) {
	if req.QuestionnaireID != nil {
		if err := s.questionnaireService.ValidateQuestionnaireExists(*req.QuestionnaireID); err != nil {
			return nil, errors.New("questionnaire not found")
		}
	}

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
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.QuestionnaireID != nil {
		if err := s.questionnaireService.ValidateQuestionnaireExists(*req.QuestionnaireID); err != nil {
			return nil, errors.New("questionnaire not found")
		}
	}

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
	complete, err := s.repo.GetComplete(id)
	if err != nil {
		return nil, err
	}
	fmt.Printf("DEBUG: Looking for consultation ID: %d\n", id)

	diagnostics, err := s.diagnosticRepo.GetByConsultationIDWithTreatments(id)
	if err != nil {
		return nil, err
	}

	complete.Diagnoses = diagnostics
	return complete, nil
}

//
// -------------------------------------------------------------
// New method: CreateFromQuestionnaire (ERD-based, uses GetQuestionnaireWithQuestions)

func (s *consultationService) CreateFromQuestionnaire(req models.ConsultationFromQuestionnaireRequest) (*models.Consultation, error) {
	if req.PatientID <= 0 {
		return nil, errors.New("patientId is required")
	}
	if req.QuestionnaireID <= 0 {
		return nil, errors.New("questionnaireId is required")
	}

	consultation := models.Consultation{
		PatientID:       req.PatientID,
		QuestionnaireID: &req.QuestionnaireID,
		Reason:          strings.TrimSpace(req.Reason),
		Date:            time.Now(),
	}
	if consultation.Reason == "" {
		consultation.Reason = "Cuestionario completado"
	}

	id, err := s.repo.Create(consultation)
	if err != nil {
		return nil, err
	}
	consultation.ID = id

	qData, err := s.questionnaireService.GetQuestionnaireWithQuestions(req.QuestionnaireID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch questionnaire structure: %w", err)
	}

	for _, q := range qData.Questions {
		ans, ok := req.Answers[strconv.Itoa(q.ID)]
		if !ok {
			continue
		}

		cq := models.ConsultationQuestion{
			ConsultationID: id,
			QuestionID:     q.ID,
			QuestionName:   q.Name,
			QuestionType:   q.Type,
			Bilateral:      q.Bilateral,
			Comment:        normalizeStringPtr(ans.Comment),
		}

		switch q.Type {
		case "entero", "float":
			if q.Bilateral {
				cq.IntValues = []int{
					toInt(normalizeValue(ans.OI)),
					toInt(normalizeValue(ans.OD)),
				}
			} else {
				cq.IntValues = []int{toInt(normalizeValue(ans.Value))}
			}

		case "bool", "booleano":
			cq.BoolValue = boolPtrFromAny(normalizeValue(ans.Value))

		case "texto", "string":
			if q.Bilateral {
				cq.TextValues = []string{
					normalizeValue(ans.OI),
					normalizeValue(ans.OD),
				}
			} else {
				cq.TextValues = []string{normalizeValue(ans.Value)}
			}

		default:
			if q.Bilateral {
				cq.TextValues = []string{
					normalizeValue(ans.OI),
					normalizeValue(ans.OD),
				}
			} else {
				cq.TextValues = []string{normalizeValue(ans.Value)}
			}
		}

		if _, err := s.repo.CreateConsultationQuestion(cq); err != nil {
			return nil, fmt.Errorf("error inserting question %d: %w", q.ID, err)
		}
	}

	return &consultation, nil
}

// ----------------- Helper Functions ------------------

func normalizeValue(v interface{}) string {
	switch t := v.(type) {
	case *string:
		if t != nil {
			return *t
		}
	case string:
		return t
	case float64:
		return fmt.Sprintf("%v", t)
	case int:
		return strconv.Itoa(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	}
	return ""
}

func normalizeStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	str := strings.TrimSpace(*s)
	if str == "" {
		return nil
	}
	return &str
}

func toInt(v string) int {
	if v == "" {
		return 0
	}
	i, _ := strconv.Atoi(v)
	return i
}

func boolPtrFromAny(v string) *bool {
	if v == "" {
		return nil
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "si", "sí", "on":
		val := true
		return &val
	case "0", "false", "no", "off":
		val := false
		return &val
	default:
		if i, err := strconv.Atoi(v); err == nil {
			val := i != 0
			return &val
		}
		return nil
	}
}
