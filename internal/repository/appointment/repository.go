package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"software-backend/internal/models"

	_ "github.com/lib/pq"
)

// Custom errors, probably gonna be moved
var ErrAppointmentNotFound = errors.New("appointment not found in repository")

// Interface for appointment data operations
type AppointmentRepository interface {
	GetAppointmentByID(id int) (*models.Appointment, error)
	CreateAppointment(appointment models.Appointment) (*models.Appointment, error)
	UpdateAppointment(appointment models.Appointment) error
	DeleteAppointment(id int) error
	ListAppointmentsInDateRange(startTime, endTime time.Time) ([]models.Appointment, error)
	HasOverlappingAppointment(start, end time.Time, excludeID *int) (bool, error)
}

// Struct to manage dependencies
type appointmentRepository struct {
	db *sql.DB
}

// Constructor to pass on dependencies
func NewAppointmentRepository(dbConn *sql.DB) AppointmentRepository {
	return &appointmentRepository{
		db: dbConn,
	}
}

// Get an appointment by ID
func (r *appointmentRepository) GetAppointmentByID(id int) (*models.Appointment, error) {
	// Build query
	query := `
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
	`

	// Create model
	appt := &models.Appointment{}
	var durationSeconds int64
	var patientName sql.NullString
	var patientID sql.NullInt64
	var fecha time.Time

	// Scan into model
	err := r.db.QueryRow(query, id).Scan(
		&appt.ID,
		&patientID,
		&patientName,
		&fecha,
		&durationSeconds,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppointmentNotFound
		}
		return nil, fmt.Errorf("repository: failed to get appointment by ID %d: %w", id, err)
	}

	// Null handling
	if patientID.Valid {
		appt.PatientID = int(patientID.Int64)
	} else {
		appt.PatientID = 0
	}
	if patientName.Valid {
		appt.Name = patientName.String
	} else {
		appt.Name = ""
	}

	// Time - Interval management, Postgres is storing a BigInt in seconds
	appt.Start = fecha

	appt.Duration = time.Duration(durationSeconds) * time.Second

	return appt, nil
}

// Create an appointment
func (r *appointmentRepository) CreateAppointment(appointment models.Appointment) (*models.Appointment, error) {
	// Build query
	query := `INSERT INTO citas (paciente_id, nombre, fecha, duracion)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id`

	// Manage nullable patientID & create appointmentID
	var appointmentID int

	var patientIDValue interface{}
	if appointment.PatientID == 0 {
		patientIDValue = nil
	} else {
		patientIDValue = appointment.PatientID
	}

	// Duration management
	durationSeconds := int64(appointment.Duration / time.Second)

	// DEBUG LOG
	log.Printf("Repository: Creating appointment. Original Start: %v, Name: %s, Duration: %v, Duration in Seconds: %d",
		appointment.Start, appointment.Name, appointment.Duration, durationSeconds)

	// Insert values into query & exec
	err := r.db.QueryRow(query,
		patientIDValue,
		appointment.Name,
		appointment.Start,
		durationSeconds,
	).Scan(&appointmentID)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to create appointment: %w", err)
	}

	appointment.ID = appointmentID

	return &appointment, nil
}

// Update an appointment
func (r *appointmentRepository) UpdateAppointment(appointment models.Appointment) error {
	// Build query
	query := `UPDATE citas SET
				paciente_id = $1,
				nombre = $2,
				fecha = $3,
				duracion = $4
			  WHERE id = $5`

	// PatientID null management
	var patientIDValue interface{}
	if appointment.PatientID == 0 {
		patientIDValue = nil
	} else {
		patientIDValue = appointment.PatientID
	}

	// Duration time / format management
	durationSeconds := int64(appointment.Duration / time.Second)

	// Pass values to query & exec
	result, err := r.db.Exec(query,
		patientIDValue,
		appointment.Name,
		appointment.Start,
		durationSeconds,
		appointment.ID,
	)
	if err != nil {
		return fmt.Errorf("repository: failed to update appointment ID %d: %w", appointment.ID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to check rows affected for update ID %d: %w", appointment.ID, err)
	}
	if rowsAffected == 0 {
		return ErrAppointmentNotFound
	}
	return nil
}

// Delete an appointment
func (r *appointmentRepository) DeleteAppointment(id int) error {
	// Build & exec query
	query := `DELETE FROM citas WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("repository: failed to delete appointment ID %d: %w", id, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to check rows affected for delete ID %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return ErrAppointmentNotFound
	}
	return nil
}

// Get appointments within a data range
func (r *appointmentRepository) ListAppointmentsInDateRange(startTime, endTime time.Time) ([]models.Appointment, error) {
	// Build query
	query := `
		SELECT
            id,
            paciente_id,
            nombre,
            fecha,
            duracion
        FROM
            citas
        WHERE
            fecha BETWEEN $1 AND $2
        ORDER BY fecha
	`

	queryStartTime := startTime
	queryEndTime := endTime

	log.Printf("Repository: Executing query for date range: %s", query)
	log.Printf("Repository: Parameter 1 ($1): Value=%v, Type=%T, Location=%v", queryStartTime, queryStartTime, queryStartTime.Location())
	log.Printf("Repository: Parameter 2 ($2): Value=%v, Type=%T, Location=%v", queryEndTime, queryEndTime, queryEndTime.Location())

	// Exec query
	rows, err := r.db.Query(query, queryStartTime, queryEndTime)
	if err != nil {
		log.Printf("Repository: Error during db.Query: %v", err)
		return nil, fmt.Errorf("repository: failed to list appointments in date range %v to %v: %w", startTime, endTime, err)
	}
	defer rows.Close()

	// Scan into appointment slice
	appointments := []models.Appointment{}
	rowCount := 0
	for rows.Next() {
		rowCount++
		var appt models.Appointment
		var durationSeconds int64
		var patientName sql.NullString
		var patientID sql.NullInt64
		var fecha time.Time

		err := rows.Scan(
			&appt.ID,
			&patientID,
			&patientName,
			&fecha,
			&durationSeconds,
		)
		if err != nil {
			log.Printf("repository: error scanning row %d: %v", rowCount, err)
			continue
		}

		if patientID.Valid {
			appt.PatientID = int(patientID.Int64)
		} else {
			appt.PatientID = 0
		}
		if patientName.Valid {
			appt.Name = patientName.String
		} else {
			appt.Name = ""
		}

		appt.Start = fecha

		appt.Duration = time.Duration(durationSeconds) * time.Second // Convert seconds to Duration

		appointments = append(appointments, appt)
	}

	if err = rows.Err(); err != nil {
		log.Printf("repository: error after iterating rows: %v", err)
		return nil, fmt.Errorf("repository: error after iterating appointment rows in date range: %w", err)
	}

	log.Printf("Repository: Successfully listed %d appointments.", len(appointments))

	// Return resulting slice
	return appointments, nil
}

// Verify if an appointment is overlapping with another
func (r *appointmentRepository) HasOverlappingAppointment(start, end time.Time, excludeID *int) (bool, error) {
	// Build query
	query := `
		SELECT 1 FROM citas
		WHERE NOT (
			(fecha + make_interval(secs => duracion)) <= $1
			OR fecha >= $2
		)
	`
	// Dynamically build args
	args := []interface{}{start, end}
	if excludeID != nil {
		query += " AND id != $3"
		args = append(args, *excludeID)
	}
	query += " LIMIT 1"

	row := r.db.QueryRow(query, args...)
	var dummy int
	err := row.Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil // No overlap
	}
	if err != nil {
		return false, err
	}
	return true, nil // Overlap found
}
