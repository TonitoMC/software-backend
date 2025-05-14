package service

import (
	"errors"
	"fmt"
	"sort" // Needed if you want to sort appointments within each day group
	"time" // Needed for time operations

	"software-backend/internal/models"     // Need Appointment model
	"software-backend/internal/repository" // Need AppointmentRepository interface
)

// Define custom errors specific to appointment service logic
var (
	ErrAppointmentConflict = errors.New("appointment time slot conflict")
	ErrInvalidAppointment  = errors.New("invalid appointment data")
	// Add other service-specific errors as needed
)

// AppointmentService defines the interface for appointment-related business logic operations.
type AppointmentService interface {
	GetAppointmentsInDateRangeAndGroupedByDay(startTime, endTime time.Time) (map[string][]models.Appointment, error)
}

// appointmentService implements the AppointmentService interface.
type appointmentService struct {
	// Dependency on the AppointmentRepository interface
	apptRepo repository.AppointmentRepository
}

// NewAppointmentService creates and returns a new instance of AppointmentService.
// It injects the AppointmentRepository dependency.
func NewAppointmentService(apptRepo repository.AppointmentRepository) AppointmentService {
	return &appointmentService{
		apptRepo: apptRepo,
	}
}

// GetAppointmentsInDateRangeAndGroupedByDay implements the service interface.
// It fetches appointments from the repository and then groups them by day.
func (s *appointmentService) GetAppointmentsInDateRangeAndGroupedByDay(startTime, endTime time.Time) (map[string][]models.Appointment, error) {
	// 1. Call the repository to get the list of appointments within the date range
	appointments, err := s.apptRepo.ListAppointmentsInDateRange(startTime, endTime) // Use the repository method
	if err != nil {
		// Handle potential repository errors (e.g., database connection issues)
		// Do NOT check for ErrAppointmentNotFound here, as returning zero appointments is valid.
		return nil, fmt.Errorf("service: failed to fetch appointments from repository: %w", err) // Wrap the error
	}

	// 2. Group the fetched appointments by the date part of their start time (YYYY-MM-DD string)
	grouped := make(map[string][]models.Appointment)
	for _, appt := range appointments {
		// Format the start time's date into a YYYY-MM-DD string as the map key
		dateKey := appt.Start.Format("2006-01-02")

		// Append the appointment to the slice for this date key
		// Append takes care of creating the slice if the key is new
		grouped[dateKey] = append(grouped[dateKey], appt)
	}

	// 3. Optional: Sort appointments within each date group by time
	// This makes the API response more predictable and organized.
	for dateStr, appts := range grouped {
		// Use a stable sort if maintaining original order among equals is important
		sort.SliceStable(appts, func(i, j int) bool {
			// Sort by the Start time
			return appts[i].Start.Before(appts[j].Start)
		})
		grouped[dateStr] = appts // Assign the sorted slice back
	}

	// 4. Return the grouped results
	return grouped, nil
}

// --- Implement other AppointmentService methods here as needed ---
// func (s *appointmentService) CreateAppointment(appointment models.Appointment) (*models.Appointment, error) { ... }
// func (s *appointmentService) GetAppointmentByID(id int) (*models.Appointment, error) { ... }
// ... and so on for all methods in the interface.
