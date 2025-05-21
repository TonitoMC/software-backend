package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"software-backend/internal/models"
	"software-backend/internal/repository"
)

// Custom errors, probably moved onto separate file in the future
var (
	ErrAppointmentConflict = errors.New("appointment time slot conflict")
	ErrInvalidAppointment  = errors.New("invalid appointment data")
)

// Interface defines methods expected from the service
type AppointmentService interface {
	GetAppointmentsInDateRangeAndGroupedByDay(startTime, endTime time.Time) (map[string][]models.Appointment, error)
	GetTodaysAppointments() (map[string][]models.Appointment, error)
	GetAppointmentsForMonth(year int, month time.Month) (map[string][]models.Appointment, error)
	GetAppointmentsForDate(date time.Time) ([]models.Appointment, error)
	DeleteAppointment(id int) error
	CreateAppointment(appointment models.Appointment) (*models.Appointment, error)
	UpdateAppointment(appointment models.Appointment) error
}

// Struct to manage dependencies
type appointmentService struct {
	apptRepo             repository.AppointmentRepository
	businessHoursService BusinessHoursService
}

// Constructor to pass on dependencies
func NewAppointmentService(apptRepo repository.AppointmentRepository, bhService BusinessHoursService) AppointmentService {
	return &appointmentService{
		apptRepo:             apptRepo,
		businessHoursService: bhService,
	}
}

// Create a new appointment
func (s *appointmentService) CreateAppointment(appointment models.Appointment) (*models.Appointment, error) {
	// Basic input validation
	if appointment.PatientID == 0 && appointment.Name == "" {
		return nil, fmt.Errorf("either patient ID or name must be provided")
	}

	// Get business hours for validation
	intervals, err := s.businessHoursService.GetBusinessHoursForDate(appointment.Start)
	if err != nil {
		return nil, fmt.Errorf("failed to get business hours: %w", err)
	}

	// Check overlap vs other scheduled appointments
	start := appointment.Start
	end := appointment.Start.Add(appointment.Duration)
	overlap, err := s.apptRepo.HasOverlappingAppointment(start, end, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check for overlapping appointments: %w", err)
	}
	// Return conflict on conflict
	if overlap {
		return nil, ErrAppointmentConflict
	}

	// Check if interval within business hours
	within, err := isWithinBusinessHours(start, end, intervals)
	if err != nil {
		return nil, fmt.Errorf("failed to parse business hours: %w", err)
	}
	// Return error on appointment outside business hours
	if !within {
		return nil, errors.New("appointment outside working hours")
	}
	return s.apptRepo.CreateAppointment(appointment)
}

// Update an appointment given the new values including ID
func (s *appointmentService) UpdateAppointment(appointment models.Appointment) error {
	// Check against overlapping appointment
	start := appointment.Start
	end := appointment.Start.Add(appointment.Duration)
	overlap, err := s.apptRepo.HasOverlappingAppointment(start, end, nil)
	if err != nil {
		return fmt.Errorf("failed to check for overlapping appointments: %w", err)
	}
	if overlap {
		// Return conflict on conflict
		return ErrAppointmentConflict
	}
	return s.apptRepo.UpdateAppointment(appointment)
}

// Get appointments in a date range, grouping them by day
func (s *appointmentService) GetAppointmentsInDateRangeAndGroupedByDay(startTime, endTime time.Time) (map[string][]models.Appointment, error) {
	// Get appointments from repository
	appointments, err := s.apptRepo.ListAppointmentsInDateRange(startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("service: failed to fetch appointments from repository: %w", err)
	}

	// Group by day in a map
	grouped := make(map[string][]models.Appointment)
	for _, appt := range appointments {
		dateKey := appt.Start.Format("2006-01-02")
		grouped[dateKey] = append(grouped[dateKey], appt)
	}
	for dateStr, appts := range grouped {
		sort.SliceStable(appts, func(i, j int) bool {
			return appts[i].Start.Before(appts[j].Start)
		})
		grouped[dateStr] = appts
	}

	// Return map with grouped appointments
	return grouped, nil
}

// Get appointments for current day using time.Now()
func (s *appointmentService) GetTodaysAppointments() (map[string][]models.Appointment, error) {
	// Get current time
	now := time.Now()

	// Generate start of day at 00:00 & end of day at 23:59
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	// Get appointments from repository & return them
	return s.GetAppointmentsInDateRangeAndGroupedByDay(startOfDay, endOfDay)
}

// Get appointments for a specific date
func (s *appointmentService) GetAppointmentsForDate(date time.Time) ([]models.Appointment, error) {
	// Generate start & end of day as 'time'
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	// Get appointments within interval from repository
	appointments, err := s.apptRepo.ListAppointmentsInDateRange(startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("service: failed to fetch appointments for date: %w", err)
	}

	sort.SliceStable(appointments, func(i, j int) bool {
		return appointments[i].Start.Before(appointments[j].Start)
	})

	return appointments, nil
}

// Get appointments for a specific month (year-month format)
func (s *appointmentService) GetAppointmentsForMonth(year int, month time.Month) (map[string][]models.Appointment, error) {
	// Generate start / end times for month interval
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Now().Location())
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	endOfMonth := startOfNextMonth.Add(-time.Nanosecond)

	// Get appointments from repository & return them
	return s.GetAppointmentsInDateRangeAndGroupedByDay(startOfMonth, endOfMonth)
}

// Delete an appointment
func (s *appointmentService) DeleteAppointment(id int) error {
	// Delete appointment via repository
	return s.apptRepo.DeleteAppointment(id)
}

// Returns true if the appointment is fully within any business interval
func isWithinBusinessHours(appointmentStart, appointmentEnd time.Time, intervals []models.BusinessHourInterval) (bool, error) {
	// Parse interval, as a day can have multiple intervals to indicate lunch break for example
	for _, interval := range intervals {
		intervalStart, err := time.ParseInLocation("15:04", interval.Start, appointmentStart.Location())
		if err != nil {
			return false, err
		}
		intervalEnd, err := time.ParseInLocation("15:04", interval.End, appointmentStart.Location())
		if err != nil {
			return false, err
		}
		// Set the date to match appointmentStart
		intervalStart = time.Date(appointmentStart.Year(), appointmentStart.Month(), appointmentStart.Day(),
			intervalStart.Hour(), intervalStart.Minute(), 0, 0, appointmentStart.Location())
		intervalEnd = time.Date(appointmentStart.Year(), appointmentStart.Month(), appointmentStart.Day(),
			intervalEnd.Hour(), intervalEnd.Minute(), 0, 0, appointmentStart.Location())

		if !appointmentStart.Before(intervalStart) && !appointmentEnd.After(intervalEnd) {
			return true, nil
		}
	}
	return false, nil
}
