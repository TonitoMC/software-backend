package exam

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

	repo := NewExamRepository(db)
	patientID := 123

	expectedType := "Blood Test"
	expectedDate := time.Now().Truncate(time.Second)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT tipo, fecha FROM examenes WHERE paciente_id = $1`,
	)).
		WithArgs(patientID).
		WillReturnRows(sqlmock.NewRows([]string{"tipo", "fecha"}).
			AddRow(expectedType, expectedDate),
		)

	exams, err := repo.GetByPatientID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exams) != 1 {
		t.Fatalf("expected 1 exam, got %d", len(exams))
	}
	if exams[0].Type != expectedType || !exams[0].Date.Equal(expectedDate) {
		t.Errorf("unexpected exam: %+v", exams[0])
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

	repo := NewExamRepository(db)
	patientID := 999

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT tipo, fecha FROM examenes WHERE paciente_id = $1`,
	)).
		WithArgs(patientID).
		WillReturnRows(sqlmock.NewRows([]string{"tipo", "fecha"}))

	exams, err := repo.GetByPatientID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exams) != 0 {
		t.Errorf("expected 0 exams, got %d", len(exams))
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

	repo := NewExamRepository(db)
	patientID := 123

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT tipo, fecha FROM examenes WHERE paciente_id = $1`,
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
