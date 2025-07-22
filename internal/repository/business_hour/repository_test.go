package businesshour

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetBusinessHoursForDate_SpecialHoursFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewBusinessHoursRepository(db)
	testDate := time.Date(2024, 7, 18, 0, 0, 0, 0, time.UTC)

	// Mock special hours query
	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT hora_apertura, hora_cierre
        FROM horarios_especiales
        WHERE fecha = $1
    `)).
		WithArgs(testDate.Format("2006-01-02")).
		WillReturnRows(sqlmock.NewRows([]string{"hora_apertura", "hora_cierre"}).
			AddRow(
				time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
				time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC),
			),
		)

	intervals, err := repo.GetBusinessHoursForDate(testDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(intervals) != 1 || intervals[0].Start != "09:00" || intervals[0].End != "17:00" {
		t.Errorf("unexpected intervals: %+v", intervals)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetBusinessHoursForDate_RegularHoursFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewBusinessHoursRepository(db)
	testDate := time.Date(2024, 7, 18, 0, 0, 0, 0, time.UTC) // Thursday

	// Mock special hours query (no rows)
	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT hora_apertura, hora_cierre
        FROM horarios_especiales
        WHERE fecha = $1
    `)).
		WithArgs(testDate.Format("2006-01-02")).
		WillReturnRows(sqlmock.NewRows([]string{"hora_apertura", "hora_cierre"}))

	// Mock regular hours query
	weekday := int(testDate.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT hora_apertura, hora_cierre
        FROM horarios_laborales
        WHERE dia_semana = $1
    `)).
		WithArgs(weekday).
		WillReturnRows(sqlmock.NewRows([]string{"hora_apertura", "hora_cierre"}).
			AddRow(
				time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
				time.Date(0, 1, 1, 16, 0, 0, 0, time.UTC),
			),
		)

	intervals, err := repo.GetBusinessHoursForDate(testDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(intervals) != 1 || intervals[0].Start != "08:00" || intervals[0].End != "16:00" {
		t.Errorf("unexpected intervals: %+v", intervals)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetBusinessHoursForDate_NoHoursFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewBusinessHoursRepository(db)
	testDate := time.Date(2024, 7, 18, 0, 0, 0, 0, time.UTC)

	// Mock special hours query (no rows)
	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT hora_apertura, hora_cierre
        FROM horarios_especiales
        WHERE fecha = $1
    `)).
		WithArgs(testDate.Format("2006-01-02")).
		WillReturnRows(sqlmock.NewRows([]string{"hora_apertura", "hora_cierre"}))

	// Mock regular hours query (no rows)
	weekday := int(testDate.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT hora_apertura, hora_cierre
        FROM horarios_laborales
        WHERE dia_semana = $1
    `)).
		WithArgs(weekday).
		WillReturnRows(sqlmock.NewRows([]string{"hora_apertura", "hora_cierre"}))

	intervals, err := repo.GetBusinessHoursForDate(testDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(intervals) != 0 {
		t.Errorf("expected no intervals, got: %+v", intervals)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetBusinessHoursForDate_SpecialHoursQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewBusinessHoursRepository(db)
	testDate := time.Date(2024, 7, 18, 0, 0, 0, 0, time.UTC)

	// Mock special hours query error
	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT hora_apertura, hora_cierre
        FROM horarios_especiales
        WHERE fecha = $1
    `)).
		WithArgs(testDate.Format("2006-01-02")).
		WillReturnError(sql.ErrConnDone)

	_, err = repo.GetBusinessHoursForDate(testDate)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
