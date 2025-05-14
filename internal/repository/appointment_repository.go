package repository

import (
	"errors"
	"sync"
	"time"

	"software-backend/internal/models"
)

// ErrAppointmentNotFound is returned when an appointment is not found by the repository.
var ErrAppointmentNotFound = errors.New("appointment not found in repository")

// AppointmentRepository defines the interface for appointment data access operations.
type AppointmentRepository interface {
	GetAppointmentByID(id int) (*models.Appointment, error)
	CreateAppointment(appointment models.Appointment) (*models.Appointment, error)
	UpdateAppointment(appointment models.Appointment) error
	DeleteAppointment(id int) error
	ListAppointmentsByPatientID(patientID int) ([]models.Appointment, error)
	ListAppointmentsInDateRange(startTime, endTime time.Time) ([]models.Appointment, error)
}

// --- Mock Appointment Repository Implementation ---

// MockAppointmentRepository is an in-memory implementation of the AppointmentRepository interface.
type MockAppointmentRepository struct {
	appointments map[int]*models.Appointment
	mu           sync.Mutex
	nextApptID   int
}

// NewMockAppointmentRepository creates a new instance of MockAppointmentRepository
// and pre-populates it with some dummy data for testing.
func NewMockAppointmentRepository() AppointmentRepository {
	initialAppointments := make(map[int]*models.Appointment)

	now := time.Now()
	initialAppointments[1] = &models.Appointment{
		ID: 1, PatientID: 123, Name: "Routine check-up", // Use Name field
		Start: now.Add(time.Hour * 24), Duration: time.Minute * 30, // Use Start field
		// Status: models.AppointmentStatusScheduled, // Status isn't in your model, removed
		// Notes: "Routine check-up", // Notes isn't in your model, removed
	}
	initialAppointments[2] = &models.Appointment{
		ID: 2, PatientID: 123, Name: "Follow-up appointment", // Use Name field
		Start: now.Add(time.Hour * 48).Add(time.Hour * 9), Duration: time.Minute * 45, // Use Start field
		// Status: models.AppointmentStatusScheduled, // Status isn't in your model, removed
		// Notes: "Follow-up appointment", // Notes isn't in your model, removed
	}
	initialAppointments[3] = &models.Appointment{
		ID: 3, PatientID: 456, Name: "Consultation Bob", // Use Name field
		Start: now.Add(time.Hour * 24).Add(time.Hour * 10), Duration: time.Minute * 20, // Use Start field
		// Status: models.AppointmentStatusCompleted, // Status isn't in your model, removed
		// Notes: "Consultation", // Notes isn't in your model, removed
	}
	initialAppointments[4] = &models.Appointment{
		ID: 4, PatientID: 123, Name: "Past Appointment", // Use Name field
		Start: now.Add(time.Hour * -12), Duration: time.Minute * 30, // Use Start field
		// Status: models.AppointmentStatusCompleted, // Status isn't in your model, removed
		// Notes: "Past appointment", // Notes isn't in your model, removed
	}

	nextID := 1
	for id := range initialAppointments {
		if id >= nextID {
			nextID = id + 1
		}
	}

	return &MockAppointmentRepository{
		appointments: initialAppointments,
		nextApptID:   nextID,
	}
}

// GetAppointmentByID implements the AppointmentRepository interface for the mock.
func (r *MockAppointmentRepository) GetAppointmentByID(id int) (*models.Appointment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	appointment, ok := r.appointments[id]
	if !ok {
		return nil, ErrAppointmentNotFound
	}
	copiedAppt := *appointment
	return &copiedAppt, nil
}

// CreateAppointment implements the AppointmentRepository interface for the mock.
func (r *MockAppointmentRepository) CreateAppointment(appointment models.Appointment) (*models.Appointment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	appointment.ID = r.nextApptID
	r.nextApptID++
	r.appointments[appointment.ID] = &appointment
	createdAppt := appointment
	return &createdAppt, nil
}

// UpdateAppointment implements the AppointmentRepository interface for the mock.
func (r *MockAppointmentRepository) UpdateAppointment(appointment models.Appointment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.appointments[appointment.ID]
	if !ok {
		return ErrAppointmentNotFound
	}
	r.appointments[appointment.ID] = &appointment
	return nil
}

// DeleteAppointment implements the AppointmentRepository interface for the mock.
func (r *MockAppointmentRepository) DeleteAppointment(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.appointments[id]
	if !ok {
		return ErrAppointmentNotFound
	}
	delete(r.appointments, id)
	return nil
}

// ListAppointmentsByPatientID implements the AppointmentRepository interface for the mock.
func (r *MockAppointmentRepository) ListAppointmentsByPatientID(patientID int) ([]models.Appointment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	appointments := []models.Appointment{}
	for _, appt := range r.appointments {
		if appt.PatientID == patientID {
			copiedAppt := *appt
			appointments = append(appointments, copiedAppt)
		}
	}
	// Optional: Sort by start time
	// sort.SliceStable(appointments, func(i, j int) bool {
	//     return appointments[i].Start.Before(appointments[j].Start)
	// })
	return appointments, nil
}

// ListAppointmentsInDateRange implements the AppointmentRepository interface for the mock.
func (r *MockAppointmentRepository) ListAppointmentsInDateRange(startTime, endTime time.Time) ([]models.Appointment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	appointments := []models.Appointment{}
	for _, appt := range r.appointments {
		// Check if appointment's start time is within the range [startTime, endTime]
		if (appt.Start.Equal(startTime) || appt.Start.After(startTime)) && (appt.Start.Equal(endTime) || appt.Start.Before(endTime)) {
			copiedAppt := *appt
			appointments = append(appointments, copiedAppt)
		}
	}
	// Optional: Sort by start time
	// sort.SliceStable(appointments, func(i, j int) bool {
	//     return appointments[i].Start.Before(appointments[j].Start)
	// })
	return appointments, nil
}

