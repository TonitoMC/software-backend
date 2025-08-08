package treatment

import (
	"database/sql"

	"software-backend/internal/models"
)

type TreatmentRepository interface {
	GetByDiagnosticID(diagnosticID int) ([]models.Treatment, error)
	Create(t models.Treatment) (int, error)
}

type treatmentRepository struct {
	db *sql.DB
}

func NewTreatmentRepository(db *sql.DB) TreatmentRepository {
	return &treatmentRepository{db: db}
}

func (r *treatmentRepository) GetByDiagnosticID(diagnosticID int) ([]models.Treatment, error) {
	query := `
		SELECT id, diagnostico_id, componente_activo, presentacion, dosificacion, frecuencia, tiempo
		FROM tratamientos
		WHERE diagnostico_id = $1
	`
	rows, err := r.db.Query(query, diagnosticID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var treatments []models.Treatment
	for rows.Next() {
		var t models.Treatment
		if err := rows.Scan(
			&t.ID,
			&t.DiagnosticID,
			&t.ActiveComponent,
			&t.Presentation,
			&t.Dosage,
			&t.Frequency,
			&t.Duration,
		); err != nil {
			return nil, err
		}
		treatments = append(treatments, t)
	}
	return treatments, nil
}

func (r *treatmentRepository) Create(t models.Treatment) (int, error) {
	query := `
		INSERT INTO tratamientos (diagnostico_id, componente_activo, presentacion, dosificacion, frecuencia, tiempo)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int
	err := r.db.QueryRow(
		query,
		t.DiagnosticID,
		t.ActiveComponent,
		t.Presentation,
		t.Dosage,
		t.Frequency,
		t.Duration,
	).Scan(&id)
	return id, err
}
