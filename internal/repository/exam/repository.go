package exam

import (
	"database/sql"
	"fmt"

	"software-backend/internal/models"
)

type ExamRepository interface {
	GetByPatientID(patientID int) ([]models.Exam, error)
	GetByID(examID int) (*models.Exam, error)
	UpdateFileMetadata(examID int, s3Key string, fileSize int64, mimeType string) error
	GetPending() ([]*models.Exam, error)
}

type examRepository struct {
	db *sql.DB
}

func NewExamRepository(db *sql.DB) ExamRepository {
	return &examRepository{db: db}
}

func (r *examRepository) GetByPatientID(patientID int) ([]models.Exam, error) {
	query := `
        SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type 
        FROM examenes 
        WHERE paciente_id = $1 
        ORDER BY fecha DESC`

	rows, err := r.db.Query(query, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exams []models.Exam
	for rows.Next() {
		var exam models.Exam
		err := rows.Scan(
			&exam.ID,
			&exam.PatientID,
			&exam.ConsultaID,
			&exam.Type,
			&exam.Date,
			&exam.S3Key,
			&exam.FileSize,
			&exam.MimeType,
		)
		if err != nil {
			return nil, err
		}

		exam.SetHasFile() // Set the computed field
		exams = append(exams, exam)
	}

	return exams, nil
}

func (r *examRepository) GetByID(examID int) (*models.Exam, error) {
	query := `
        SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type 
        FROM examenes 
        WHERE id = $1`

	var exam models.Exam
	err := r.db.QueryRow(query, examID).Scan(
		&exam.ID,
		&exam.PatientID,
		&exam.ConsultaID,
		&exam.Type,
		&exam.Date,
		&exam.S3Key,
		&exam.FileSize,
		&exam.MimeType,
	)
	if err != nil {
		return nil, err
	}

	exam.SetHasFile()
	return &exam, nil
}

func (r *examRepository) GetPending() ([]*models.Exam, error) {
	query := `SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type
						FROM examenes
						WHERE s3_key IS NULL
						ORDER BY fecha ASC;
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	var exams []*models.Exam
	for rows.Next() {
		exam := &models.Exam{}

		err := rows.Scan(
			&exam.ID,
			&exam.PatientID,
			&exam.ConsultaID,
			&exam.Type,
			&exam.Date,
			&exam.S3Key,
			&exam.FileSize,
			&exam.MimeType,
		)
		if err != nil {
			fmt.Printf("Error scanning row for pending exam: %v\n", err)
			return nil, err
		}
		exam.SetHasFile()
		exams = append(exams, exam)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return exams, nil
}

func (r *examRepository) UpdateFileMetadata(examID int, s3Key string, fileSize int64, mimeType string) error {
	query := `
        UPDATE examenes 
        SET s3_key = $1, file_size = $2, mime_type = $3 
        WHERE id = $4`

	_, err := r.db.Exec(query, s3Key, fileSize, mimeType, examID)
	return err
}
