package services

import (
	"software-backend/internal/models"
	"software-backend/internal/repository/questionnaire"
)

type QuestionnaireService interface {
	GetActiveQuestionnaires() ([]models.Questionnaire, error)
	GetQuestionnaireWithQuestions(id int) (*models.QuestionnaireWithQuestions, error)
	ValidateQuestionnaireExists(id int) error
	GetAllQuestionnaires() ([]models.Questionnaire, error)
	UpdateQuestionnaire(id int, questionnaire *models.QuestionnaireUpdate) error
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

// Add these methods to the existing questionnaireService struct

func (s *questionnaireService) GetAllQuestionnaires() ([]models.Questionnaire, error) {
	return s.questionnaireRepo.GetAll()
}

func (s *questionnaireService) UpdateQuestionnaire(id int, questionnaire *models.QuestionnaireUpdate) error {
	// First validate the questionnaire exists
	if err := s.ValidateQuestionnaireExists(id); err != nil {
		return err
	}
	return s.questionnaireRepo.Update(id, questionnaire)
}

func (s *questionnaireService) SetQuestionnaireActive(id int, active bool) error {
	// First validate the questionnaire exists
	if err := s.ValidateQuestionnaireExists(id); err != nil {
		return err
	}
	return s.questionnaireRepo.SetActive(id, active)
}
