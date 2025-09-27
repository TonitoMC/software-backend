package patient

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"software-backend/internal/models"

	_ "github.com/lib/pq"
)

// Custom errors, like others is probably going to be moved
var ErrPatientNotFound = errors.New("patient not found in repository")

// Interface for interaction with repository
type PatientRepository interface {
	GetPatientByID(id int) (*models.Patient, error)
	CreatePatient(patient models.Patient) (*models.Patient, error)
	UpdatePatient(patient models.Patient) error
	DeletePatient(id int) error
	ListPatients() ([]models.Patient, error)
	SearchPatients(query string, limit int) ([]models.Patient, error)
}

// Struct to pass on dependencies
// note: called sqlPatientRepository as there was previously a mock repository,
// TODO: change sqlPatientRepository to patientRepository
type sqlPatientRepository struct {
	db *sql.DB
}

// Constructor to pass on dependencies
func NewPatientRepository(dbConn *sql.DB) PatientRepository {
	return &sqlPatientRepository{db: dbConn}
}

// Get a patient & their record via ID
func (r *sqlPatientRepository) GetPatientByID(id int) (*models.Patient, error) {
	// Build query
	query := `
		SELECT
            p.id,
            p.nombre,
            p.fecha_nacimiento, -- Date of birth
            p.telefono,
            p.sexo,
            a.medicos, -- Antecedentes fields can be NULL due to LEFT JOIN
            a.familiares,
            a.oculares,
            a.alergicos,
            a.otros
        FROM
            pacientes p
        LEFT JOIN
            antecedentes a ON p.id = a.paciente_id
        WHERE
            p.id = $1
	`

	// Create patient model
	patient := &models.Patient{}
	var dateOfBirth time.Time
	var phone sql.NullString
	var antecedenteMedical sql.NullString
	var antecedenteFamily sql.NullString
	var antecedenteOcular sql.NullString
	var antecedenteAlergic sql.NullString
	var antecedenteOther sql.NullString

	// Scan into patient
	err := r.db.QueryRow(query, id).Scan(
		&patient.ID,
		&patient.Name,
		&dateOfBirth,
		&phone,
		&patient.Sex,
		&antecedenteMedical,
		&antecedenteFamily,
		&antecedenteOcular,
		&antecedenteAlergic,
		&antecedenteOther,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPatientNotFound
		}
		return nil, fmt.Errorf("repository: failed to get patient by ID %d: %w", id, err)
	}

	patient.DateOfBirth = dateOfBirth
	if phone.Valid {
		patient.Phone = phone.String
	} else {
		patient.Phone = ""
	}

	// Handle antecedentes
	if antecedenteMedical.Valid || antecedenteFamily.Valid || antecedenteOcular.Valid || antecedenteAlergic.Valid || antecedenteOther.Valid {
		patient.Antecedentes = &models.Antecedentes{
			Medical: antecedenteMedical.String,
			Family:  antecedenteFamily.String,
			Ocular:  antecedenteOcular.String,
			Alergic: antecedenteAlergic.String,
			Other:   antecedenteOther.String,
		}
	} else {
		patient.Antecedentes = nil
	}

	return patient, nil
}

// Create a new patient via Patient model
func (r *sqlPatientRepository) CreatePatient(patient models.Patient) (*models.Patient, error) {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("repository: failed to begin transaction for creating patient: %w", err)
	}
	// Rollback if error
	defer tx.Rollback()

	// Insert into pacientes table
	patientQuery := `
		INSERT INTO pacientes (nombre, fecha_nacimiento, telefono, sexo)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var patientID int
	phoneValue := sql.NullString{String: patient.Phone, Valid: patient.Phone != ""}

	// Get patient's ID
	err = tx.QueryRow(patientQuery, patient.Name, patient.DateOfBirth, phoneValue, patient.Sex).Scan(&patientID)
	if err != nil {
		// Check for duplicate name/other constraints if applicable
		return nil, fmt.Errorf("repository: failed to create patient in pacientes table: %w", err)
	}

	// Update the patient model with the generated ID
	patient.ID = patientID

	// Insert related antecedentes if provided
	if patient.Antecedentes != nil {
		antecedentesQuery := `
			INSERT INTO antecedentes (paciente_id, medicos, familares, oculares, alergicos, otros)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.Exec(antecedentesQuery,
			patient.ID,
			patient.Antecedentes.Medical,
			patient.Antecedentes.Family,
			patient.Antecedentes.Ocular,
			patient.Antecedentes.Alergic,
			patient.Antecedentes.Other,
		)
		if err != nil {
			log.Printf("repository: failed to create antecedentes for patient %d: %v", patient.ID, err)
			return nil, fmt.Errorf("repository: failed to create antecedentes for patient %d: %w", patient.ID, err)
		}
	}

	// Commit transaction if both inserts successful
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: failed to commit transaction for creating patient: %w", err)
	}

	return &patient, nil
}

// Update a patient via Patient model with ID
func (r *sqlPatientRepository) UpdatePatient(patient models.Patient) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("repository: failed to begin transaction for updating patient: %w", err)
	}
	// Defer rollback in case of failure
	defer tx.Rollback()

	// Update pacientes table
	patientUpdateQuery := `
		UPDATE pacientes
		SET nombre = $1, fecha_nacimiento = $2, telefono = $3, sexo = $4
		WHERE id = $5
	`

	// Handle nullable rows & possible errors
	phoneValue := sql.NullString{String: patient.Phone, Valid: patient.Phone != ""}
	result, err := tx.Exec(patientUpdateQuery, patient.Name, patient.DateOfBirth, phoneValue, patient.Sex, patient.ID)
	if err != nil {
		return fmt.Errorf("repository: failed to update patient in pacientes table (ID %d): %w", patient.ID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to check rows affected for patient update (ID %d): %w", patient.ID, err)
	}
	if rowsAffected == 0 {
		return ErrPatientNotFound
	}

	// Handle antecedentes updates
	if patient.Antecedentes != nil {
		// Check if Antecedentes already exist for this patient
		var existingAntecedentesID int
		checkQuery := `SELECT id FROM antecedentes WHERE paciente_id = $1`
		err := tx.QueryRow(checkQuery, patient.ID).Scan(&existingAntecedentesID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Antecedentes do not exist, INSERT
				antecedentesInsertQuery := `
					INSERT INTO antecedentes (paciente_id, medicos, familiares, oculares, alergicos, otros)
					VALUES ($1, $2, $3, $4, $5, $6)
				`
				_, err := tx.Exec(antecedentesInsertQuery,
					patient.ID,
					patient.Antecedentes.Medical,
					patient.Antecedentes.Family,
					patient.Antecedentes.Ocular,
					patient.Antecedentes.Alergic,
					patient.Antecedentes.Other,
				)
				if err != nil {
					log.Printf("repository: failed to insert antecedentes for patient %d: %v", patient.ID, err)
					return fmt.Errorf("repository: failed to insert antecedentes for patient %d: %w", patient.ID, err)
				}
			} else {
				// Some other error checking for existing Antecedentes
				log.Printf("repository: failed to check for existing antecedentes for patient %d: %v", patient.ID, err)
				return fmt.Errorf("repository: failed to check for existing antecedentes for patient %d: %w", patient.ID, err)
			}
		} else {
			// Antecedentes exist, UPDATE
			antecedentesUpdateQuery := `
				UPDATE antecedentes
				SET medicos = $1, familiares = $2, oculares = $3, alergicos = $4, otros = $5
				WHERE paciente_id = $6
			`
			_, err := tx.Exec(antecedentesUpdateQuery,
				patient.Antecedentes.Medical,
				patient.Antecedentes.Family,
				patient.Antecedentes.Ocular,
				patient.Antecedentes.Alergic,
				patient.Antecedentes.Other,
				patient.ID,
			)
			if err != nil {
				log.Printf("repository: failed to update antecedentes for patient %d: %v", patient.ID, err)
				return fmt.Errorf("repository: failed to update antecedentes for patient %d: %w", patient.ID, err)
			}
		}
	} else {
		// TODO pending implementation, has to do with patient creation / update flow. Leaving for when that
		// module is implemented
	}

	// Commit transaction if queries successful
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: failed to commit transaction for updating patient: %w", err)
	}

	return nil
}

// Delete a patient based on ID
func (r *sqlPatientRepository) DeletePatient(id int) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("repository: failed to begin transaction for deleting patient: %w", err)
	}
	// Defer rollback in case of failure
	defer tx.Rollback()

	// Delete related antecedentes
	antecedentesDeleteQuery := `DELETE FROM antecedentes WHERE paciente_id = $1`
	_, err = tx.Exec(antecedentesDeleteQuery, id)
	if err != nil {
		log.Printf("repository: failed to delete antecedentes for patient %d: %v", id, err)
	}

	// Delete paciente
	patientDeleteQuery := `DELETE FROM pacientes WHERE id = $1`
	result, err := tx.Exec(patientDeleteQuery, id)
	if err != nil {
		return fmt.Errorf("repository: failed to delete patient from pacientes table (ID %d): %w", id, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to check rows affected for patient delete (ID %d): %w", id, err)
	}
	if rowsAffected == 0 {
		return ErrPatientNotFound
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: failed to commit transaction for deleting patient: %w", err)
	}

	return nil
}

// Lists patients
func (r *sqlPatientRepository) ListPatients() ([]models.Patient, error) {
	// Query for patient details only
	query := `
		SELECT
            id,
            nombre,
            fecha_nacimiento,
            telefono,
            sexo
        FROM
            pacientes
		ORDER BY nombre -- Order alphabetically by name
	`

	// Exec query w/o params, not needed
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to list patients: %w", err)
	}
	defer rows.Close()

	// Scan into patient list
	patients := []models.Patient{}
	for rows.Next() {
		var patient models.Patient
		var dateOfBirth time.Time
		var phone sql.NullString

		err := rows.Scan(
			&patient.ID,
			&patient.Name,
			&dateOfBirth,
			&phone,
			&patient.Sex,
		)
		if err != nil {
			log.Printf("repository: error scanning patient row: %v", err)
			continue
		}

		patient.DateOfBirth = dateOfBirth
		if phone.Valid {
			patient.Phone = phone.String
		} else {
			patient.Phone = ""
		}

		patients = append(patients, patient)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: error after iterating patient rows: %w", err)
	}

	return patients, nil
}

// Get a list of patients of names fuzzy-match
func (r *sqlPatientRepository) SearchPatients(query string, limit int) ([]models.Patient, error) {
	// Search by name (case-insensitive, partial match)
	sqlQuery := `
        SELECT id, nombre, fecha_nacimiento, telefono, sexo
        FROM pacientes
        WHERE nombre ILIKE $1
        ORDER BY nombre
        LIMIT $2
    `
	// Exec query
	rows, err := r.db.Query(sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to search patients: %w", err)
	}
	defer rows.Close()

	// Scan into patient list
	var patients []models.Patient
	for rows.Next() {
		var patient models.Patient
		var dateOfBirth time.Time
		var phone sql.NullString

		err := rows.Scan(
			&patient.ID,
			&patient.Name,
			&dateOfBirth,
			&phone,
			&patient.Sex,
		)
		if err != nil {
			continue
		}
		patient.DateOfBirth = dateOfBirth
		if phone.Valid {
			patient.Phone = phone.String
		} else {
			patient.Phone = ""
		}
		patients = append(patients, patient)
	}
	// Return resulting list
	return patients, nil
}
