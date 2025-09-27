package appointment

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetAppointmentByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewAppointmentRepository(db)

	// Prepare expected data
	expectedID := 1
	expectedPatientID := int64(42)
	expectedName := "John Doe"
	expectedFecha := time.Now().Truncate(time.Second)
	expectedDuration := int64(3600)

	// Set up mock
	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT
            id,
            paciente_id,
            nombre,
            fecha,
            duracion
        FROM
            citas
        WHERE
            id = $1
    `)).
		WithArgs(expectedID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "paciente_id", "nombre", "fecha", "duracion",
		}).AddRow(
			expectedID,
			expectedPatientID,
			expectedName,
			expectedFecha,
			expectedDuration,
		))

	// Call method
	appt, err := repo.GetAppointmentByID(expectedID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if appt.ID != expectedID ||
		appt.PatientID != int(expectedPatientID) ||
		appt.Name != expectedName ||
		!appt.Start.Equal(expectedFecha) ||
		appt.Duration != time.Duration(expectedDuration)*time.Second {
		t.Errorf("unexpected appointment: %+v", appt)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetAppointmentByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewAppointmentRepository(db)

	expectedID := 999

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT
            id,
            paciente_id,
            nombre,
            fecha,
            duracion
        FROM
            citas
        WHERE
            id = $1
    `)).
		WithArgs(expectedID).
		WillReturnError(sql.ErrNoRows)

	appt, err := repo.GetAppointmentByID(expectedID)
	if err != ErrAppointmentNotFound {
		t.Fatalf("expected ErrAppointmentNotFound, got: %v", err)
	}
	if appt != nil {
		t.Errorf("expected nil appointment, got: %+v", appt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
