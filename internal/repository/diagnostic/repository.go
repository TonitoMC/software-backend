package diagnostic

import (
	"database/sql"

	"software-backend/internal/models"
)

type DiagnosticRepository interface {
	GetByConsultationIDWithTreatments(consultationID int) ([]models.Diagnostic, error)
	CreateBatch(consultationID int, diagnostics []models.Diagnostic) error
}

type diagnosticRepository struct {
	db *sql.DB
}

func NewDiagnosticRepository(db *sql.DB) DiagnosticRepository {
	return &diagnosticRepository{db: db}
}

func (r *diagnosticRepository) GetByConsultationIDWithTreatments(consultationID int) ([]models.Diagnostic, error) {
	query := `
		SELECT d.id, d.nombre, d.recomendacion,
		       t.id, t.componente_activo, t.presentacion, t.dosificacion, t.frecuencia, t.tiempo
		FROM diagnosticos d
		LEFT JOIN tratamientos t ON d.id = t.diagnostico_id
		WHERE d.consulta_id = $1
		ORDER BY d.id, t.id
	`
	rows, err := r.db.Query(query, consultationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diagnostics []models.Diagnostic
	diagnosticMap := make(map[int]*models.Diagnostic)

	for rows.Next() {
		var (
			diagID          int
			name            string
			recommendation  string
			treatID         sql.NullInt64
			activeComponent sql.NullString
			presentation    sql.NullString
			dosage          sql.NullString
			frequency       sql.NullString
			duration        sql.NullString
		)

		if err := rows.Scan(
			&diagID, &name, &recommendation,
			&treatID, &activeComponent, &presentation, &dosage, &frequency, &duration,
		); err != nil {
			return nil, err
		}

		diag, exists := diagnosticMap[diagID]
		if !exists {
			diag = &models.Diagnostic{
				ID:             diagID,
				Name:           name,
				Recommendation: recommendation,
				ConsultationID: consultationID,
				Treatments:     []models.Treatment{},
			}
			diagnosticMap[diagID] = diag
			diagnostics = append(diagnostics, *diag)
		}

		if treatID.Valid {
			diag.Treatments = append(diag.Treatments, models.Treatment{
				ID:              int(treatID.Int64),
				DiagnosticID:    diagID,
				ActiveComponent: activeComponent.String,
				Presentation:    presentation.String,
				Dosage:          dosage.String,
				Frequency:       frequency.String,
				Duration:        duration.String,
			})
		}
	}

	return diagnostics, nil
}

func (r *diagnosticRepository) CreateBatch(consultationID int, diagnostics []models.Diagnostic) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, d := range diagnostics {
		var diagID int
		err := tx.QueryRow(
			`INSERT INTO diagnosticos (nombre, recomendacion, consulta_id)
			 VALUES ($1, $2, $3) RETURNING id`,
			d.Name, d.Recommendation, consultationID,
		).Scan(&diagID)
		if err != nil {
			return err
		}

		for _, t := range d.Treatments {
			_, err := tx.Exec(
				`INSERT INTO tratamientos
				 (diagnostico_id, componente_activo, presentacion, dosificacion, frecuencia, tiempo)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				diagID, t.ActiveComponent, t.Presentation, t.Dosage, t.Frequency, t.Duration,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
