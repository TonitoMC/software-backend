package repository

import (
	"database/sql"
	"time"

	"software-backend/internal/models"
)

type ConsultationRepository interface {
	GetByPatientID(patientID int) ([]models.Consultation, error)
}

type consultationRepository struct {
	db *sql.DB
}

func NewConsultationRepository(db *sql.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) GetByPatientID(patientID int) ([]models.Consultation, error) {
	consultations := []models.Consultation{}
	query := `SELECT motivo, fecha FROM consultas WHERE paciente_id = $1`
	rows, err := r.db.Query(query, patientID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var consultation models.Consultation
		var date time.Time

		err := rows.Scan(
			&consultation.Reason,
			&date,
		)
		if err != nil {
			return nil, err
		}
		consultation.Date = date

		consultations = append(consultations, consultation)
	}
	return consultations, nil
}
