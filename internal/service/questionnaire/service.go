package services

import (
	"software-backend/internal/models"
	"software-backend/internal/repository/questionnaire"
)

type QuestionnaireService interface {
	GetActiveQuestionnaires() ([]models.Questionnaire, error)
	GetQuestionnaireWithQuestions(id int) (*models.QuestionnaireWithQuestions, error)
	ValidateQuestionnaireExists(id int) error
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
