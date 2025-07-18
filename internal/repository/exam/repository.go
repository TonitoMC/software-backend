package exam

import (
	"database/sql"
	"time"

	"software-backend/internal/models"
)

// Interface defines methods to interact with repository
type ExamRepository interface {
	GetByPatientID(patientID int) ([]models.Exam, error)
}

// Struct to manage dependencies
type examRepository struct {
	db *sql.DB
}

// Constructor to pass on dependencies
func NewExamRepository(db *sql.DB) ExamRepository {
	return &examRepository{db: db}
}

func (r *examRepository) GetByPatientID(patientID int) ([]models.Exam, error) {
	exams := []models.Exam{}
	query := `SELECT tipo, fecha FROM examenes WHERE paciente_id = $1`
	rows, err := r.db.Query(query, patientID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var exam models.Exam
		var date time.Time

		err := rows.Scan(
			&exam.Type,
			&date,
		)
		if err != nil {
			return nil, err
		}
		exam.Date = date

		exams = append(exams, exam)
	}
	return exams, nil
}
