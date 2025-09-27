package questionnaire

import (
	"database/sql"

	"software-backend/internal/models"
)

type QuestionnaireRepository interface {
	GetActive() ([]models.Questionnaire, error)
	GetWithQuestions(id int) (*models.QuestionnaireWithQuestions, error)
	GetByID(id int) (*models.Questionnaire, error)
	GetAll() ([]models.Questionnaire, error)
	Update(id int, questionnaire *models.QuestionnaireUpdate) error
	SetActive(id int, active bool) error
}

type questionnaireRepository struct {
	db *sql.DB
}

func NewQuestionnaireRepository(db *sql.DB) QuestionnaireRepository {
	return &questionnaireRepository{db: db}
}

func (r *questionnaireRepository) GetActive() ([]models.Questionnaire, error) {
	query := `
		SELECT id, nombre, version, activo
		FROM cuestionarios 
		WHERE activo = true
		ORDER BY nombre, version DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questionnaires []models.Questionnaire
	for rows.Next() {
		var q models.Questionnaire
		err := rows.Scan(
			&q.ID,
			&q.Name,
			&q.Version,
			&q.Active,
		)
		if err != nil {
			return nil, err
		}
		questionnaires = append(questionnaires, q)
	}
	return questionnaires, nil
}

func (r *questionnaireRepository) GetByID(id int) (*models.Questionnaire, error) {
	query := `
		SELECT id, nombre, version, activo
		FROM cuestionarios 
		WHERE id = $1`

	var q models.Questionnaire
	err := r.db.QueryRow(query, id).Scan(
		&q.ID,
		&q.Name,
		&q.Version,
		&q.Active,
	)
	if err != nil {
		return nil, err
	}

	return &q, nil
}

func (r *questionnaireRepository) GetWithQuestions(id int) (*models.QuestionnaireWithQuestions, error) {
	// First get the questionnaire
	questionnaire, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Then get questions in order
	query := `
		SELECT p.id, p.nombre, p.tipo, p.bilateral, pc.orden
		FROM preguntas p
		INNER JOIN preguntas_cuestionarios pc ON p.id = pc.pregunta_id
		WHERE pc.cuestionario_id = $1
		ORDER BY pc.orden`

	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.QuestionWithOrder
	for rows.Next() {
		var q models.QuestionWithOrder
		err := rows.Scan(
			&q.ID,
			&q.Name,
			&q.Type,
			&q.Bilateral,
			&q.Order,
		)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	return &models.QuestionnaireWithQuestions{
		Questionnaire: *questionnaire,
		Questions:     questions,
	}, nil
}

// Add these methods to the existing questionnaireRepository struct

func (r *questionnaireRepository) GetAll() ([]models.Questionnaire, error) {
	query := `
		SELECT id, nombre, version, activo
		FROM cuestionarios 
		ORDER BY nombre, version DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questionnaires []models.Questionnaire
	for rows.Next() {
		var q models.Questionnaire
		err := rows.Scan(
			&q.ID,
			&q.Name,
			&q.Version,
			&q.Active,
		)
		if err != nil {
			return nil, err
		}
		questionnaires = append(questionnaires, q)
	}
	return questionnaires, nil
}

func (r *questionnaireRepository) Update(id int, questionnaire *models.QuestionnaireUpdate) error {
	query := `
		UPDATE cuestionarios 
		SET nombre = $2, version = $3
		WHERE id = $1`

	_, err := r.db.Exec(query, id, questionnaire.Name, questionnaire.Version)
	return err
}

func (r *questionnaireRepository) SetActive(id int, active bool) error {
	query := `
		UPDATE cuestionarios 
		SET activo = $2
		WHERE id = $1`

	_, err := r.db.Exec(query, id, active)
	return err
}
