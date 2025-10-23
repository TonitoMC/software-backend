package consultation

import (
	"database/sql"

	"software-backend/internal/models"

	"github.com/lib/pq"
)

type ConsultationRepository interface {
	Create(consultation models.Consultation) (int, error)
	GetByID(id int) (*models.Consultation, error)
	GetByPatientID(patientID int) ([]models.Consultation, error)
	Update(id int, consultation models.Consultation) error
	Delete(id int) error
	GetComplete(id int) (*models.CompleteConsultation, error)

	// New method for saving questionnaire answers
	CreateConsultationQuestion(q models.ConsultationQuestion) (int, error)
}

type consultationRepository struct {
	db *sql.DB
}

func NewConsultationRepository(db *sql.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) Create(consultation models.Consultation) (int, error) {
	query := `
		INSERT INTO consultas (paciente_id, cuestionario_id, motivo, fecha)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var id int
	err := r.db.QueryRow(
		query,
		consultation.PatientID,
		consultation.QuestionnaireID,
		consultation.Reason,
		consultation.Date,
	).Scan(&id)

	return id, err
}

func (r *consultationRepository) GetByID(id int) (*models.Consultation, error) {
	query := `
		SELECT id, paciente_id, cuestionario_id, motivo, fecha 
		FROM consultas 
		WHERE id = $1`

	var c models.Consultation
	var questionnaireID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&c.ID,
		&c.PatientID,
		&questionnaireID,
		&c.Reason,
		&c.Date,
	)
	if err != nil {
		return nil, err
	}

	if questionnaireID.Valid {
		qID := int(questionnaireID.Int64)
		c.QuestionnaireID = &qID
	}

	return &c, nil
}

func (r *consultationRepository) GetByPatientID(patientID int) ([]models.Consultation, error) {
	query := `
		SELECT id, paciente_id, cuestionario_id, motivo, fecha 
		FROM consultas 
		WHERE paciente_id = $1 
		ORDER BY fecha DESC`

	rows, err := r.db.Query(query, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var consultations []models.Consultation
	for rows.Next() {
		var c models.Consultation
		var questionnaireID sql.NullInt64

		err := rows.Scan(
			&c.ID,
			&c.PatientID,
			&questionnaireID,
			&c.Reason,
			&c.Date,
		)
		if err != nil {
			return nil, err
		}

		if questionnaireID.Valid {
			qID := int(questionnaireID.Int64)
			c.QuestionnaireID = &qID
		}

		consultations = append(consultations, c)
	}
	return consultations, nil
}

func (r *consultationRepository) Update(id int, consultation models.Consultation) error {
	query := `
		UPDATE consultas 
		SET motivo = $2, fecha = $3, cuestionario_id = $4
		WHERE id = $1`

	_, err := r.db.Exec(
		query,
		id,
		consultation.Reason,
		consultation.Date,
		consultation.QuestionnaireID,
	)

	return err
}

func (r *consultationRepository) Delete(id int) error {
	query := `DELETE FROM consultas WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *consultationRepository) GetComplete(id int) (*models.CompleteConsultation, error) {
	// Get basic consultation
	consultation, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Get questionnaire answers

	query := `
    SELECT 
        cp.id, cp.consulta_id, cp.pregunta_id,
        p.nombre AS question_name,
        p.tipo AS question_type,
        p.bilateral AS question_bilateral,
        cp.valores_textos, cp.valores_enteros, cp.valores_booleanos,
        cp.valor_texto, cp.valor_entero, cp.valor_booleano, cp.comentario
    FROM consultas_preguntas cp
    JOIN preguntas p ON p.id = cp.pregunta_id
    WHERE cp.consulta_id = $1
    ORDER BY p.id`

	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.ConsultationQuestion
	for rows.Next() {
		var q models.ConsultationQuestion

		var textValues pq.StringArray
		var intValues pq.Int64Array
		var boolValues pq.BoolArray

		err := rows.Scan(
			&q.ID, &q.ConsultationID, &q.QuestionID,
			&q.QuestionName, &q.QuestionType, &q.Bilateral,
			&textValues, &intValues, &boolValues,
			&q.TextValue, &q.IntValue, &q.BoolValue, &q.Comment,
		)
		if err != nil {
			return nil, err
		}

		q.TextValues = []string(textValues)

		for _, v := range intValues {
			q.IntValues = append(q.IntValues, int(v))
		}

		for _, v := range boolValues {
			q.BoolValues = append(q.BoolValues, bool(v))
		}

		questions = append(questions, q)
	}

	return &models.CompleteConsultation{
		Consultation: *consultation,
		Questions:    questions,
	}, nil
}

func (r *consultationRepository) CreateConsultationQuestion(q models.ConsultationQuestion) (int, error) {
	query := `
		INSERT INTO consultas_preguntas 
			(consulta_id, pregunta_id, valores_textos, valores_enteros, valor_booleano, comentario)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id int
	err := r.db.QueryRow(
		query,
		q.ConsultationID,
		q.QuestionID,
		pq.StringArray(q.TextValues), // ✅ properly stores text arrays
		pq.Int64Array(intSliceToInt64(q.IntValues)), // ✅ properly stores integer arrays
		q.BoolValue, // ✅ single bool value (can be nil)
		q.Comment,   // ✅ single comment (can be nil)
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// helper to convert []int → []int64 for pq.Int64Array
func intSliceToInt64(src []int) []int64 {
	res := make([]int64, len(src))
	for i, v := range src {
		res[i] = int64(v)
	}
	return res
}
