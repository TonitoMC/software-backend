package service

import (
	"fmt"

	"software-backend/internal/models"
	"software-backend/internal/repository/questionnaire"
)

type QuestionnaireService interface {
	GetActiveQuestionnaires() ([]models.Questionnaire, error)
	GetQuestionnaireWithQuestions(id int) (*models.QuestionnaireWithQuestions, error)
	ValidateQuestionnaireExists(id int) error
	GetAllQuestionnaires() ([]models.Questionnaire, error)
	UpdateQuestionnaire(id int, questionnaire *models.QuestionnaireWithQuestions) error
	SetQuestionnaireActive(id int, active bool) error
}

type questionnaireService struct {
	questionnaireRepo questionnaire.QuestionnaireRepository
}

func NewQuestionnaireService(questionnaireRepo questionnaire.QuestionnaireRepository) QuestionnaireService {
	return &questionnaireService{
		questionnaireRepo: questionnaireRepo,
	}
}

func (s *questionnaireService) GetActiveQuestionnaires() ([]models.Questionnaire, error) {
	return s.questionnaireRepo.GetActive()
}

func (s *questionnaireService) GetQuestionnaireWithQuestions(id int) (*models.QuestionnaireWithQuestions, error) {
	return s.questionnaireRepo.GetWithQuestions(id)
}

func (s *questionnaireService) ValidateQuestionnaireExists(id int) error {
	_, err := s.questionnaireRepo.GetByID(id)
	return err
}

func (s *questionnaireService) GetAllQuestionnaires() ([]models.Questionnaire, error) {
	return s.questionnaireRepo.GetAll()
}

func (s *questionnaireService) UpdateQuestionnaire(id int, questionnaire *models.QuestionnaireWithQuestions) error {
	// 1️⃣ Ensure old exists
	if err := s.ValidateQuestionnaireExists(id); err != nil {
		return err
	}

	// 2️⃣ Create a brand new questionnaire row
	newID, err := s.questionnaireRepo.Create(&models.Questionnaire{
		Name:    questionnaire.Name,
		Version: questionnaire.Version,
		Active:  questionnaire.Active,
	})
	if err != nil {
		return fmt.Errorf("failed to create new version: %w", err)
	}

	// 3️⃣ Deactivate the previous version
	if err := s.questionnaireRepo.SetActive(id, false); err != nil {
		return fmt.Errorf("failed to deactivate old version: %w", err)
	}

	// 4️⃣ Ensure all questions are reattached cleanly to the new questionnaire
	cleanQuestions := make([]models.QuestionWithOrder, 0, len(questionnaire.Questions))
	for _, q := range questionnaire.Questions {
		cleanQuestions = append(cleanQuestions, models.QuestionWithOrder{
			Question: models.Question{
				ID:        q.ID, // keep same question IDs (preguntas.id)
				Name:      q.Name,
				Type:      q.Type,
				Bilateral: q.Bilateral,
			},
			Order: q.Order,
		})
	}

	// 5️⃣ Attach them using the new questionnaire ID
	if err := s.questionnaireRepo.AttachQuestions(newID, cleanQuestions); err != nil {
		return fmt.Errorf("failed to attach questions: %w", err)
	}

	return nil
}

func (s *questionnaireService) SetQuestionnaireActive(id int, active bool) error {
	return s.questionnaireRepo.SetActive(id, active)
}
