package exam

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetByPatientID_Found(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewExamRepository(db)
	patientID := 123
	examDate := time.Now().Truncate(time.Second)

	query := regexp.QuoteMeta(`
        SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type 
        FROM examenes 
        WHERE paciente_id = $1 
        ORDER BY fecha DESC`)

	mock.ExpectQuery(query).
		WithArgs(patientID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "paciente_id", "consulta_id", "tipo", "fecha", "s3_key", "file_size", "mime_type",
		}).AddRow(1, patientID, 10, "Blood", examDate, "key.pdf", 2048, "application/pdf"))

	exams, err := repo.GetByPatientID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exams) != 1 {
		t.Fatalf("expected 1 exam, got %d", len(exams))
	}
	if exams[0].Type != "Blood" || !exams[0].HasFile {
		t.Errorf("unexpected exam result: %+v", exams[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetByPatientID_Empty(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewExamRepository(db)

	query := regexp.QuoteMeta(`
        SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type 
        FROM examenes 
        WHERE paciente_id = $1 
        ORDER BY fecha DESC`)

	mock.ExpectQuery(query).
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "paciente_id", "consulta_id", "tipo", "fecha", "s3_key", "file_size", "mime_type",
		}))

	exams, err := repo.GetByPatientID(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exams) != 0 {
		t.Errorf("expected 0 exams, got %d", len(exams))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetByPatientID_QueryError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewExamRepository(db)

	query := regexp.QuoteMeta(`
        SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type 
        FROM examenes 
        WHERE paciente_id = $1 
        ORDER BY fecha DESC`)

	mock.ExpectQuery(query).
		WithArgs(123).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetByPatientID(123)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetByID_Found(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewExamRepository(db)
	examDate := time.Now().Truncate(time.Second)

	query := regexp.QuoteMeta(`
        SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type 
        FROM examenes 
        WHERE id = $1`)

	mock.ExpectQuery(query).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "paciente_id", "consulta_id", "tipo", "fecha", "s3_key", "file_size", "mime_type",
		}).AddRow(1, 123, 456, "X-Ray", examDate, "xray.pdf", 1234, "application/pdf"))

	exam, err := repo.GetByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exam.ID != 1 || exam.Type != "X-Ray" || !exam.HasFile {
		t.Errorf("unexpected exam: %+v", exam)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetPending(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewExamRepository(db)
	query := regexp.QuoteMeta(`
        SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type
        FROM examenes
        WHERE s3_key IS NULL
        ORDER BY fecha ASC;
	`)

	mock.ExpectQuery(query).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "paciente_id", "consulta_id", "tipo", "fecha", "s3_key", "file_size", "mime_type",
		}).AddRow(1, 101, 201, "CT", time.Now().Truncate(time.Second), nil, 0, "image/jpeg"))

	results, err := repo.GetPending()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].HasFile {
		t.Errorf("unexpected pending exam: %+v", results[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateFileMetadata(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewExamRepository(db)
	query := regexp.QuoteMeta(`
        UPDATE examenes 
        SET s3_key = $1, file_size = $2, mime_type = $3 
        WHERE id = $4`)

	mock.ExpectExec(query).
		WithArgs("upload/file.pdf", int64(2048), "application/pdf", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateFileMetadata(10, "upload/file.pdf", 2048, "application/pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
