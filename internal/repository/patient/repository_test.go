package patient

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetPatientByID_FoundWithAntecedentes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewPatientRepository(db)
	patientID := 1
	expectedName := "Jane Doe"
	expectedDOB := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedPhone := "555-1234"
	expectedSex := "F"
	expectedMedical := "Asthma"
	expectedFamily := "Diabetes"
	expectedOcular := "Myopia"
	expectedAlergic := "Penicillin"
	expectedOther := "None"

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(patientID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "nombre", "fecha_nacimiento", "telefono", "sexo",
			"medicos", "familiares", "oculares", "alergicos", "otros",
		}).AddRow(
			patientID,
			expectedName,
			expectedDOB,
			expectedPhone,
			expectedSex,
			expectedMedical,
			expectedFamily,
			expectedOcular,
			expectedAlergic,
			expectedOther,
		))

	patient, err := repo.GetPatientByID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patient.ID != patientID ||
		patient.Name != expectedName ||
		!patient.DateOfBirth.Equal(expectedDOB) ||
		patient.Phone != expectedPhone ||
		patient.Sex != expectedSex {
		t.Errorf("unexpected patient: %+v", patient)
	}
	if patient.Antecedentes == nil ||
		patient.Antecedentes.Medical != expectedMedical ||
		patient.Antecedentes.Family != expectedFamily ||
		patient.Antecedentes.Ocular != expectedOcular ||
		patient.Antecedentes.Alergic != expectedAlergic ||
		patient.Antecedentes.Other != expectedOther {
		t.Errorf("unexpected antecedentes: %+v", patient.Antecedentes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetPatientByID_FoundWithoutAntecedentes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewPatientRepository(db)
	patientID := 2
	expectedName := "John Smith"
	expectedDOB := time.Date(1985, 5, 5, 0, 0, 0, 0, time.UTC)
	expectedPhone := "555-5678"
	expectedSex := "M"

	// All antecedentes fields NULL
	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(patientID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "nombre", "fecha_nacimiento", "telefono", "sexo",
			"medicos", "familiares", "oculares", "alergicos", "otros",
		}).AddRow(
			patientID,
			expectedName,
			expectedDOB,
			expectedPhone,
			expectedSex,
			nil, nil, nil, nil, nil,
		))

	patient, err := repo.GetPatientByID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patient.ID != patientID ||
		patient.Name != expectedName ||
		!patient.DateOfBirth.Equal(expectedDOB) ||
		patient.Phone != expectedPhone ||
		patient.Sex != expectedSex {
		t.Errorf("unexpected patient: %+v", patient)
	}
	if patient.Antecedentes != nil {
		t.Errorf("expected nil antecedentes, got: %+v", patient.Antecedentes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetPatientByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewPatientRepository(db)
	patientID := 999

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(patientID).
		WillReturnError(sql.ErrNoRows)

	patient, err := repo.GetPatientByID(patientID)
	if !errors.Is(err, ErrPatientNotFound) {
		t.Fatalf("expected ErrPatientNotFound, got: %v", err)
	}
	if patient != nil {
		t.Errorf("expected nil patient, got: %+v", patient)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetPatientByID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewPatientRepository(db)
	patientID := 123

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(patientID).
		WillReturnError(sql.ErrConnDone)

	_, err = repo.GetPatientByID(patientID)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
