package appointment

import (
	"errors"
	"testing"
	"time"

	"software-backend/internal/mocks"
	"software-backend/internal/models"

	"github.com/golang/mock/gomock"
)

// Hand-written mock for BusinessHoursService
type mockBusinessHoursService struct {
	GetBusinessHoursForDateFunc func(date time.Time) ([]models.BusinessHourInterval, error)
}

func (m *mockBusinessHoursService) GetBusinessHoursForDate(date time.Time) ([]models.BusinessHourInterval, error) {
	return m.GetBusinessHoursForDateFunc(date)
}

func TestCreateAppointment_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAppointmentRepository(ctrl)
	bhService := &mockBusinessHoursService{
		GetBusinessHoursForDateFunc: func(date time.Time) ([]models.BusinessHourInterval, error) {
			return []models.BusinessHourInterval{{Start: "09:00", End: "17:00"}}, nil
		},
	}

	// Set up expectations
	mockRepo.EXPECT().
		HasOverlappingAppointment(gomock.Any(), gomock.Any(), gomock.Nil()).
		Return(true, nil)

	svc := NewAppointmentService(mockRepo, bhService)

	appt := models.Appointment{
		PatientID: 1,
		Name:      "Test",
		Start:     time.Date(2024, 7, 18, 10, 0, 0, 0, time.UTC),
		Duration:  time.Hour,
	}
	_, err := svc.CreateAppointment(appt)
	if !errors.Is(err, ErrAppointmentConflict) {
		t.Errorf("expected conflict error, got %v", err)
	}
}

func TestCreateAppointment_OutsideBusinessHours(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAppointmentRepository(ctrl)
	bhService := &mockBusinessHoursService{
		GetBusinessHoursForDateFunc: func(date time.Time) ([]models.BusinessHourInterval, error) {
			return []models.BusinessHourInterval{{Start: "09:00", End: "17:00"}}, nil
		},
	}

	// No overlap
	mockRepo.EXPECT().
		HasOverlappingAppointment(gomock.Any(), gomock.Any(), gomock.Nil()).
		Return(false, nil)

	svc := NewAppointmentService(mockRepo, bhService)

	appt := models.Appointment{
		PatientID: 1,
		Name:      "Test",
		Start:     time.Date(2024, 7, 18, 18, 0, 0, 0, time.UTC), // 6pm, outside business hours
		Duration:  time.Hour,
	}
	_, err := svc.CreateAppointment(appt)
	if err == nil || err.Error() != "appointment outside working hours" {
		t.Errorf("expected outside working hours error, got %v", err)
	}
}

func TestCreateAppointment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAppointmentRepository(ctrl)
	bhService := &mockBusinessHoursService{
		GetBusinessHoursForDateFunc: func(date time.Time) ([]models.BusinessHourInterval, error) {
			return []models.BusinessHourInterval{{Start: "09:00", End: "17:00"}}, nil
		},
	}

	// No overlap
	mockRepo.EXPECT().
		HasOverlappingAppointment(gomock.Any(), gomock.Any(), gomock.Nil()).
		Return(false, nil)
	// CreateAppointment returns the appointment with ID set
	mockRepo.EXPECT().
		CreateAppointment(gomock.Any()).
		DoAndReturn(func(appt models.Appointment) (*models.Appointment, error) {
			appt.ID = 42
			return &appt, nil
		})

	svc := NewAppointmentService(mockRepo, bhService)

	appt := models.Appointment{
		PatientID: 1,
		Name:      "Test",
		Start:     time.Date(2024, 7, 18, 10, 0, 0, 0, time.UTC),
		Duration:  time.Hour,
	}
	created, err := svc.CreateAppointment(appt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 42 {
		t.Errorf("expected ID 42, got %d", created.ID)
	}
}

func TestCreateAppointment_InvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAppointmentRepository(ctrl)
	bhService := &mockBusinessHoursService{}

	svc := NewAppointmentService(mockRepo, bhService)

	appt := models.Appointment{
		PatientID: 0,
		Name:      "",
		Start:     time.Now(),
		Duration:  time.Hour,
	}
	_, err := svc.CreateAppointment(appt)
	if err == nil || err.Error() != "either patient ID or name must be provided" {
		t.Errorf("expected input validation error, got %v", err)
	}
}

func TestCreateAppointment_BusinessHoursError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAppointmentRepository(ctrl)
	bhService := &mockBusinessHoursService{
		GetBusinessHoursForDateFunc: func(date time.Time) ([]models.BusinessHourInterval, error) {
			return nil, errors.New("db error")
		},
	}

	svc := NewAppointmentService(mockRepo, bhService)

	appt := models.Appointment{
		PatientID: 1,
		Name:      "Test",
		Start:     time.Now(),
		Duration:  time.Hour,
	}
	_, err := svc.CreateAppointment(appt)
	if err == nil || err.Error() != "failed to get business hours: db error" {
		t.Errorf("expected business hours error, got %v", err)
	}
}

