package consultation

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetByPatientID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewConsultationRepository(db)
	patientID := 123

	// Prepare expected data
	expectedReason := "Routine Checkup"
	expectedDate := time.Now().Truncate(time.Second)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT motivo, fecha FROM consultas WHERE paciente_id = $1`,
	)).
		WithArgs(patientID).
		WillReturnRows(sqlmock.NewRows([]string{"motivo", "fecha"}).
			AddRow(expectedReason, expectedDate),
		)

	consultations, err := repo.GetByPatientID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(consultations) != 1 {
		t.Fatalf("expected 1 consultation, got %d", len(consultations))
	}
	if consultations[0].Reason != expectedReason || !consultations[0].Date.Equal(expectedDate) {
		t.Errorf("unexpected consultation: %+v", consultations[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetByPatientID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewConsultationRepository(db)
	patientID := 999

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT motivo, fecha FROM consultas WHERE paciente_id = $1`,
	)).
		WithArgs(patientID).
		WillReturnRows(sqlmock.NewRows([]string{"motivo", "fecha"}))

	consultations, err := repo.GetByPatientID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(consultations) != 0 {
		t.Errorf("expected 0 consultations, got %d", len(consultations))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetByPatientID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewConsultationRepository(db)
	patientID := 123

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT motivo, fecha FROM consultas WHERE paciente_id = $1`,
	)).
		WithArgs(patientID).
		WillReturnError(sql.ErrConnDone)

	_, err = repo.GetByPatientID(patientID)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
